package oss

import (
	"bytes"
	"context"
	"io"
	"sync"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/buzhiyun/go-utils/log"
)

type uploadReq struct {
	partNumber int32
	data       *[]byte //  ****************** 这里不用指针大概率会翻车
}

// 上传文件的io.WriterCloser 接口
type multipartUploadWriter struct {
	bucketName string
	objectName string
	partNumber int32
	uploadId   string
	parts      []oss.UploadPart
	uploadChan chan uploadReq
	mu         sync.Mutex
	bufferMu   sync.Mutex
	threads    int
	wg         sync.WaitGroup
	// finishChan chan interface{}
	buffer bytes.Buffer
}

func (uw *multipartUploadWriter) newPartNumber() int32 {
	uw.mu.Lock()
	defer uw.mu.Unlock()
	uw.partNumber += 1
	_partNumber := uw.partNumber
	return _partNumber

}

/*
*

  - 分片上传Writer实例
*/
func NewMultipartUploadWriter(bucketName, objectName string, threadCount int) (uw *multipartUploadWriter, err error) {
	// 初始化分片上传
	// 定义上传ID
	var uploadId string

	// 初始化分片上传请求
	initRequest := &oss.InitiateMultipartUploadRequest{
		Bucket: oss.Ptr(bucketName),
		Key:    oss.Ptr(objectName),
	}

	initResult, err := ossClient().InitiateMultipartUpload(context.TODO(), initRequest)
	if err != nil {
		log.Errorf("[UploadWriter] 初始化分片上传请求异常, %v", err)
		return nil, err
	}

	// 打印初始化分片上传的结果
	log.Infof("[UploadWriter] 初始化分片上传请求结果: %#v", *initResult.UploadId)
	uploadId = *initResult.UploadId

	_uw := &multipartUploadWriter{
		bucketName: bucketName,
		objectName: objectName,
		partNumber: 0,
		uploadId:   uploadId,
	}

	_uw.uploadChan = make(chan uploadReq, threadCount*2)
	// _uw.finishChan = make(chan any, threadCount)
	_uw.threads = threadCount

	// 启动上传线程
	for i := 0; i < threadCount; i++ {
		_uw.wg.Add(1)
		go func(threadId int) {
			c := ossClient()
			for {
				select {
				case req, ok := <-_uw.uploadChan:
					if !ok {
						log.Infof("[UploadWriter] 上传队列已关闭, uploadThread-%v 已退出", threadId)
						_uw.wg.Done()
						// _uw.finishChan <- 1
						return
					}
					_uw.uploadPart(&req, threadId, c)
				}
			}
		}(i)
	}

	return _uw, nil
}

func (uw *multipartUploadWriter) Write(p []byte) (n int, err error) {
	log.Debugf("[UploadWriter] 写入数据到缓存 %v", len(p))

	uw.bufferMu.Lock()
	// 这里不能直接丢入到队列，oss对分片的最小单元有限制，这里先写入buffer
	_, err = uw.buffer.Write(p)
	uw.bufferMu.Unlock()
	if err != nil {
		log.Errorf("[UploadWriter] 写入数据到缓冲区异常, %v", err)
		return 0, err
	}

	// 如果缓存大于1M，则写入oss
	if uw.buffer.Len() > 1024*1024 {
		log.Debugf("[UploadWriter] 缓存大于1M，开始写入oss")
		uw.flushBuffer()
	}

	// 写入分片上传队列
	return len(p), nil
}

/*
*
  - 上传分片
*/
func (uw *multipartUploadWriter) uploadPart(req *uploadReq, threadId int, client ...*oss.Client) {
	// 上传分片
	partNumber := req.partNumber
	data := *req.data

	// 初始化分片上传请求
	partRequest := &oss.UploadPartRequest{
		Bucket:     oss.Ptr(uw.bucketName), // 目标存储空间名称
		Key:        oss.Ptr(uw.objectName), // 目标对象名称
		UploadId:   oss.Ptr(uw.uploadId),   // 上传ID
		PartNumber: int32(partNumber),      // 分片编号
		Body:       bytes.NewReader(data),  // 分片内容
	}
	// log.Infof("part header %s", strings.ToUpper(hex.EncodeToString(data[:40])))
	log.Infof("[UploadWriter] uploadThread-%v 上传分片 %d , size %v", threadId, partNumber, len(data))

	partResult, err := ossClient(client...).UploadPart(context.TODO(), partRequest, func(o *oss.Options) {
		o.RetryMaxAttempts = oss.Ptr(3)
	})

	if err != nil {
		log.Errorf("[UploadWriter] 上传分片失败 %d: %v", partNumber, err)
	}

	// 记录分片上传结果
	part := oss.UploadPart{
		PartNumber: partRequest.PartNumber,
		ETag:       partResult.ETag,
	}
	uw.mu.Lock()
	uw.parts = append(uw.parts, part)
	uw.mu.Unlock()
}

func (uw *multipartUploadWriter) Close() {
	// 关闭上传队列
	uw.Finished()
}

/*
*
  - 缓存写入队列
*/
func (uw *multipartUploadWriter) flushBuffer() {
	// 剩下的内容直接写入oss
	uw.bufferMu.Lock()
	defer uw.bufferMu.Unlock()
	for uw.buffer.Len() > 1024*1024 {
		var body = make([]byte, 1024*1024)
		io.ReadFull(&uw.buffer, body)
		// body := uw.buffer.Next(1024 * 1024)
		// log.Infof("flush buffer %s", strings.ToUpper(hex.EncodeToString(body[:40])))
		uw.uploadChan <- uploadReq{
			partNumber: uw.newPartNumber(),
			data:       &body,
		}
	}
}

/*
*
  - 关闭上传队列
*/
func (uw *multipartUploadWriter) Finished() {
	uw.bufferMu.Lock()
	data := uw.buffer.Bytes()
	uw.uploadChan <- uploadReq{
		partNumber: uw.newPartNumber(),
		data:       &data,
	}
	uw.bufferMu.Unlock()

	// 关闭上传队列    ****************** 这里关闭channel可能有问题，待确认
	close(uw.uploadChan)

	// 等待上传线程完成
	uw.wg.Wait()

	// 完成分片上传请求
	request := &oss.CompleteMultipartUploadRequest{
		Bucket:   oss.Ptr(uw.bucketName),
		Key:      oss.Ptr(uw.objectName),
		UploadId: oss.Ptr(uw.uploadId),
		CompleteMultipartUpload: &oss.CompleteMultipartUpload{
			Parts: uw.parts,
		},
	}
	result, err := ossClient().CompleteMultipartUpload(context.TODO(), request)
	if err != nil {
		log.Error("[UploadWriter] 结束分片上传请求异常, %v", err)
		return
	}

	// 打印完成分片上传的结果
	log.Infof("[UploadWriter] 结束分片上传请求结果 key: %s , Status: %s", *result.Key, result.Status)

}
