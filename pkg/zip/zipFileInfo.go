package zip

import (
	"io/fs"
	"strings"
	"time"

	oss2 "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/buzhiyun/oss-package/pkg/oss"
)

/**
 * 把oss obj的文件信息转化成 fs.FileInfo
 */

type zipFileInfo struct {
	// fs.FileInfo
	isDir   bool
	name    string
	size    int64
	modTime time.Time
	mode    fs.FileMode
}

// IsDir implements fs.FileInfo.
func (zi zipFileInfo) IsDir() bool {
	return zi.isDir
}

// ModTime implements fs.FileInfo.
func (zi zipFileInfo) ModTime() time.Time {
	return zi.modTime
}

// Mode implements fs.FileInfo.
func (zi zipFileInfo) Mode() fs.FileMode {
	return zi.mode
}

func (zi zipFileInfo) Name() string {
	return zi.name
}

// Size implements fs.FileInfo.
func (zi zipFileInfo) Size() int64 {
	return zi.size
}

// Sys implements fs.FileInfo.
func (zi zipFileInfo) Sys() any {
	return nil
}

func NewZipFileInfo(isDir bool, name string, size int64, modTime time.Time, mode fs.FileMode) *zipFileInfo {
	return &zipFileInfo{
		isDir:   isDir,
		name:    name,
		size:    size,
		modTime: modTime,
		mode:    mode,
	}
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
	return zipFileInfo{
		name:    name,
		isDir:   isDir,
		size:    obj.Size,
		modTime: oss2.ToTime(obj.LastModified),
	}
}

func getFileInfoFromOssFileProperties(info oss.OssFileProperties, zipFileName string) fs.FileInfo {
	var isDir = false
	if strings.HasSuffix(zipFileName, "/") {
		isDir = true
	}
	return zipFileInfo{
		name:    zipFileName,
		isDir:   isDir,
		size:    info.Length,
		modTime: *info.ModTime,
	}

}

type GetZipfileInfo interface {
	ListFileInfo(handle func(*SrcFileProperties))
}
