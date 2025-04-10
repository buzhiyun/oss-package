package oss

import (
	"context"
	"io"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/buzhiyun/go-utils/log"
)

/*
*
  - 获取文件
*/
func GetFileSimple(bucketName, objectName string, client ...*oss.Client) (data []byte, err error) {

	// 创建获取对象的请求
	request := &oss.GetObjectRequest{
		Bucket: oss.Ptr(bucketName), // 存储空间名称
		Key:    oss.Ptr(objectName), // 对象名称
	}

	// 执行获取对象的操作并处理结果
	result, err := ossClient(client...).GetObject(context.TODO(), request, func(o *oss.Options) {
		o.RetryMaxAttempts = oss.Ptr(15)
	})
	if err != nil {
		log.Errorf("下载obj异常, %v", err)
	}
	defer result.Body.Close() // 确保在函数结束时关闭响应体
	// 一次性读取整个文件内容
	data, err = io.ReadAll(result.Body)
	if err != nil {
		log.Errorf("读取obj异常, %v", err)
	}
	return
}

/*
*
  - 获取文件
*/
func GetObjReader(bucketName, objectName string) (reader io.ReadCloser, err error) {
	// 创建获取对象的请求
	request := &oss.GetObjectRequest{
		Bucket: oss.Ptr(bucketName), // 存储空间名称
		Key:    oss.Ptr(objectName), // 对象名称
	}

	// 执行获取对象的操作并处理结果
	log.Debugf("获取obj %v", objectName)
	result, err := ossClient().GetObject(context.TODO(), request)
	if err != nil {
		log.Errorf("下载obj异常, %v", err)
	}
	return result.Body, nil
}

func GetObjMeta(bucketName, objectName string, client ...*oss.Client) (result *oss.GetObjectMetaResult, err error) {
	// 创建获取对象元数据的请求
	request := &oss.GetObjectMetaRequest{
		Bucket: oss.Ptr(bucketName), // 存储空间名称
		Key:    oss.Ptr(objectName), // 对象名称
	}

	// 执行获取对象元数据的操作并处理结果
	result, err = ossClient(client...).GetObjectMeta(context.TODO(), request)
	if err != nil {
		log.Errorf("获取 object meta 异常 %s/%s , %v", err)
	}

	return
}

func ListPathWithHandle(bucketName, prefix string, handle func(obj oss.ObjectProperties)) {
	// 创建列出对象的请求
	request := &oss.ListObjectsV2Request{
		Bucket:  oss.Ptr(bucketName),
		Prefix:  oss.Ptr(prefix), // 列举指定目录下的所有对象
		MaxKeys: 1000,
	}

	// 创建分页器
	p := ossClient().NewListObjectsV2Paginator(request)

	// 初始化页码计数器
	var i int

	// 遍历分页器中的每一页
	for p.HasNext() {
		i++
		log.Infof("获取第 %v 页的文件列表", i)

		// 获取下一页的数据
		page, err := p.NextPage(context.TODO(), func(o *oss.Options) {
			o.RetryMaxAttempts = oss.Ptr(3)
		})
		if err != nil {
			log.Errorf("获取第 %v 页数据失败, %v", i, err)
			return
		}

		//打印continue token
		// log.Debugf("ContinuationToken:%v\n", oss.ToString(page.ContinuationToken))
		// 打印该页中的每个对象的信息
		for _, obj := range page.Contents {
			// log.Infof("Object: %v, %v, %v\n", oss.ToString(obj.Key), obj.Size, oss.ToTime(obj.LastModified))
			handle(obj)
		}
	}

}

func SignObj(bucketName, objectName string, expire time.Duration, client ...*oss.Client) (url string) {
	result, err := ossClient(client...).Presign(context.TODO(), &oss.GetObjectRequest{
		Bucket: oss.Ptr(bucketName),
		Key:    oss.Ptr(objectName),
	},
		oss.PresignExpires(expire),
	)
	if err != nil {
		log.Errorf("[oss] 获取签名失败, %v", err)
	} else {
		url = result.URL
	}
	return
}
