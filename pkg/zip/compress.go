package zip

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"io/fs"
	"strings"
	"sync"

	"github.com/buzhiyun/oss-package/pkg/oss"

	oss2 "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/buzhiyun/go-utils/log"
)

type srcOssFile struct {
	fileInfo *fs.FileInfo
	objKey   *string
	data     *[]byte
}

type zipOssToOss struct {
	srcBucketName       string
	srcPrefix           string
	zipBucketName       string
	zipFileKey          string
	downloadThreadCount int                        // 下载线程数
	uploadThreadCount   int                        // 上传线程数
	downloadChan        chan oss2.ObjectProperties // 下载队列
	zipChan             chan srcOssFile            // 压缩队列
	wg                  sync.WaitGroup
	zipWg               sync.WaitGroup
}

/**
 * 初始化oss打包实例
 */
func NewZipOssToOss(ossPrefix string, zipFileKey string, downloadThreadCount, uploadThreadCount int) (z *zipOssToOss, err error) {
	_path := strings.SplitN(ossPrefix, "/", 2)
	_zippath := strings.SplitN(zipFileKey, "/", 2)

	if downloadThreadCount == 0 || uploadThreadCount == 0 {
		err = errors.New("[zip] 下载线程数或上传线程数不能为0")
		log.Fatal(err)
		return
	}

	if len(_path) == 2 && len(_zippath) == 2 {
		srcBucketName := _path[0]
		srcPrefix := _path[1]

		zipBucketName := _zippath[0]
		zipKey := _zippath[1]
		return &zipOssToOss{
			srcBucketName:       srcBucketName,
			srcPrefix:           srcPrefix,
			zipBucketName:       zipBucketName,
			zipFileKey:          zipKey,
			downloadThreadCount: downloadThreadCount,
			uploadThreadCount:   uploadThreadCount,
			downloadChan:        make(chan oss2.ObjectProperties, downloadThreadCount*2),
			zipChan:             make(chan srcOssFile, downloadThreadCount*2),
		}, nil
	}
	err = errors.New("[zip] 压缩文件路径错误")
	log.Fatal(err)
	return
}

/**
 * 下载oss文件
 */
func (z *zipOssToOss) downloadOssObj(obj oss2.ObjectProperties, dl *oss.ObjDownloader) {
	// log.Infof("[zip] 下载文件")
	data, err := dl.Download(z.srcBucketName, *obj.Key)
	if err != nil {
		log.Errorf("[zip] downloadThread-%v 下载文件失败", dl.GetId())
		return
	}
	// 丢入压缩队列
	fi := getFileInfo(obj)
	z.zipChan <- srcOssFile{
		data:     &data,
		objKey:   obj.Key,
		fileInfo: &fi,
	}
}

/**
 * 压缩文件
 */
func (z *zipOssToOss) Zip() {
	zipfile, err := oss.NewMultipartUploadWriter(z.zipBucketName, z.zipFileKey, z.uploadThreadCount)
	if err != nil {
		return
	}
	defer zipfile.Close()
	// 打开：zip文件
	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	for i := range z.downloadThreadCount {
		log.Infof("[zip] downloadThread-%v 开始下载文件", i)
		// 启动下载线程方法
		z.wg.Add(1)
		go func(threadId int) {
			dl := oss.NewObjDownloader(threadId)
			for {
				select {
				case srcFile, ok := <-z.downloadChan:
					if !ok {
						z.wg.Done()
						log.Infof("[zip] downloadThread-%v 下载文件任务完成", i)
						return
					}
					z.downloadOssObj(srcFile, dl)
				}
			}
		}(i)
	}

	// 列举oss目录下的文件
	z.wg.Add(1)
	go func() {
		oss.ListPathWithHandle(z.srcBucketName, z.srcPrefix, func(obj oss2.ObjectProperties) {
			log.Debugf("获取到 %s", *obj.Key)
			z.downloadChan <- obj
		})
		z.wg.Done()
		// 关闭下载队列，让下载线程正常退出
		close(z.downloadChan)
	}()

	// 压缩线程
	log.Info("启动压缩线程")
	z.zipWg.Add(1)
	go func() {
		for {
			select {
			case srcFile, ok := <-z.zipChan:
				if !ok {
					log.Infof("[zip] 压缩文件任务完成")
					z.zipWg.Done()
					return
				}

				log.Debugf("[zip] 压缩 %s", *srcFile.objKey)
				fi := *srcFile.fileInfo
				header, _ := zip.FileInfoHeader(fi)

				// 这地方要改
				// header.Name = fi.Name()
				header.Name = strings.TrimPrefix(*srcFile.objKey, z.srcPrefix)
				header.Name = strings.TrimPrefix(header.Name, "/")
				// 判断：文件是不是文件夹
				if !fi.IsDir() {
					// 设置：zip的文件压缩算法
					header.Method = zip.Deflate
				} else if header.Name == "" {
					break
				}

				// 创建：压缩包头部信息
				writer, err := archive.CreateHeader(header)
				if err != nil {
					log.Errorf("[zip] 创建 %s 异常: %v", header.Name, err)
					break
				}
				if !fi.IsDir() {
					// 读取obj
					br := bytes.NewReader(*srcFile.data)
					io.Copy(writer, br)
				}

			}
		}
	}()

	// 等待下载线程退出
	z.wg.Wait()
	close(z.zipChan)

	// 等待压缩线程退出
	z.zipWg.Wait()

}
