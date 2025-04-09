package zip

import oss2 "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"

type GetZipfileInfo interface {
	ListFileInfo(handle func(*oss2.ObjectProperties))
}
