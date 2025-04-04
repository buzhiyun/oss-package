package oss

import (
	"context"
	"io"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
)

/*
*
**** 方法尚未完成

  - 上传文件

  - @param reader 文件流

  - @param bucketName 存储空间名称

  - @param objectName 对象名称
*/
func Upload(reader io.Reader, bucketName, objectName string) {
	// 创建上传对象的请求
	putRequest := &oss.PutObjectRequest{
		Bucket:       oss.Ptr(bucketName),      // 存储空间名称
		Key:          oss.Ptr(objectName),      // 对象名称
		StorageClass: oss.StorageClassStandard, // 指定对象的存储类型为标准存储
		Acl:          oss.ObjectACLPrivate,     // 指定对象的访问权限为私有访问
		// Metadata: map[string]string{
		// 	"yourMetadataKey1": "yourMetadataValue1", // 设置对象的元数据
		// },
		Body: reader,
	}
	ossClient().PutObject(context.TODO(), putRequest)
}
