package zip

import (
	"io/fs"
	"strings"
	"time"

	oss2 "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
)

/**
 * 把oss obj的文件信息转化成 fs.FileInfo
 */

type ossObjInfo struct {
	// fs.FileInfo
	isDir   bool
	name    string
	size    int64
	modTime time.Time
}

// IsDir implements fs.FileInfo.
func (oi ossObjInfo) IsDir() bool {
	return oi.isDir
}

// ModTime implements fs.FileInfo.
func (oi ossObjInfo) ModTime() time.Time {
	return oi.modTime
}

// Mode implements fs.FileInfo.
func (oi ossObjInfo) Mode() fs.FileMode {
	return fs.ModePerm
}

func (oi ossObjInfo) Name() string {
	return oi.name
}

// Size implements fs.FileInfo.
func (oi ossObjInfo) Size() int64 {
	return oi.size
}

// Sys implements fs.FileInfo.
func (oi ossObjInfo) Sys() any {
	return nil
}

func getFileInfo(obj oss2.ObjectProperties) fs.FileInfo {
	key := oss2.ToString(obj.Key)
	_key := strings.Split(key, "/")
	var isDir = false
	var name string
	if _key[len(_key)-1] == "" {
		isDir = true
		if len(_key) > 1 {
			name = _key[len(_key)-2]
		} else {
			name = _key[len(_key)-1]
		}
		// log.Infof("key: %s", _key[len(_key)-2])
	} else {
		name = _key[len(_key)-1]
	}
	return ossObjInfo{
		name:    name,
		isDir:   isDir,
		size:    obj.Size,
		modTime: oss2.ToTime(obj.LastModified),
	}
}
