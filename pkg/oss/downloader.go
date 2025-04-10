package oss

import (
	"context"
	"fmt"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/buzhiyun/go-utils/log"
)

type ObjDownloader struct {
	id     int
	client *oss.Client
}

func (d *ObjDownloader) GetId() string {
	return fmt.Sprintf("%v", d.id)
}

func NewObjDownloader(id int) *ObjDownloader {
	return &ObjDownloader{
		id:     id,
		client: ossClient(),
	}
}

func (d *ObjDownloader) Download(bucketName string, objPath string) (data []byte, err error) {
	data, err = GetFileSimple(bucketName, objPath, d.client)
	return
}

func (d *ObjDownloader) GetObjMeta(bucketName, objectName string) (result *oss.GetObjectMetaResult, err error) {
	// 创建获取对象元数据的请求
	request := &oss.GetObjectMetaRequest{
		Bucket: oss.Ptr(bucketName), // 存储空间名称
		Key:    oss.Ptr(objectName), // 对象名称
	}

	// 执行获取对象元数据的操作并处理结果
	result, err = ossClient(d.client).GetObjectMeta(context.TODO(), request)
	if err != nil {
		log.Errorf("获取 object meta 异常 %s/%s , %v", err)
	}
	return
}
