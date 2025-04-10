package oss

import "strings"

func GetOssKeyFromLongKey(objectLongKey string) (bucketName, ossKey string, ok bool) {
	_key := strings.SplitN(objectLongKey, "/", 2)
	if len(_key) != 2 {
		ok = false
		return
	}
	bucketName = _key[0]
	ossKey = _key[1]
	ok = true
	return
}
