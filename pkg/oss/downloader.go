package oss

import (
	"fmt"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
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
