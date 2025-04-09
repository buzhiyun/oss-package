package fx

import (
	"bufio"
	"fmt"
	"strings"

	oss2 "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"

	"github.com/buzhiyun/go-utils/log"
	"github.com/buzhiyun/oss-package/pkg/oss"
	"github.com/buzhiyun/oss-package/pkg/zip"
)

type FxZipFileInfo struct {
	zip.GetZipfileInfo
	// zipInfoTxt   string
	examGuid     string
	templateGuid string
	zipFileName  string
}

func (fx *FxZipFileInfo) ListFileInfo(handle func(fileInfo *oss2.ObjectProperties)) {
	zipFile := fmt.Sprintf("arithmeticcenterNew/TempFile/%s/%s-tmp/zip-file.txt", fx.examGuid, fx.templateGuid)
	// oss 中文件位置
	zipInfoTxt, err := oss.GetObjReader("yjreport", zipFile)
	if err != nil {
		log.Errorf("[fx] 获取zip-file.txt 失败%s, %s", zipFile, err)
		return
	}
	defer zipInfoTxt.Close()

	// 创建一个 Scanner 来逐行读取文件
	scanner := bufio.NewScanner(zipInfoTxt)

	// 逐行读取文件内容
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		// log.Infof("第 %d 行: %s", lineNumber, scanner.Text())
		// 处理文件每一行
		if lineNumber == 1 {
			fx.zipFileName = scanner.Text()
			log.Infof("[fx] zip文件 %s", fx.zipFileName)
			continue
		} else {
			line := strings.SplitN(scanner.Text(), "|", 2)
			if len(line) < 2 {
				log.Errorf("[fx] 分隔符异常, 第 %d 行: %s", lineNumber, scanner.Text())
				continue
			}
			_, ossKey := line[0], line[1]

			_objKey := strings.SplitN(ossKey, "/", 2)
			if len(_objKey) != 2 {
				log.Errorf("[fx] oss文件异常 第 %d 行: %s", lineNumber, scanner.Text())
				continue
			}

			srcBucket, srcKey := _objKey[0], _objKey[1]

			// 以下内容可能会有点慢，这部分 需要优化多线程
			// 获取文件meta
			result, err := oss.GetObjMeta(srcBucket, srcKey)
			if err != nil {
				log.Errorf("[fx] 获取oss meta 失败 %s, %s", srcKey, err)
				continue
			}

			handle(&oss2.ObjectProperties{
				Key:          &srcKey,
				LastModified: result.LastModified,
				Size:         result.ContentLength,
			})
		}

	}

	// 检查是否在读取过程中发生错误
	if err := scanner.Err(); err != nil {
		log.Infof("读取文件时发生错误:", err)
	}

}

func (fx *FxZipFileInfo) GetZipFileName() string {
	if fx.zipFileName == "" {
		zipFile := fmt.Sprintf("arithmeticcenterNew/TempFile/%s/%s-tmp/zip-file.txt", fx.examGuid, fx.templateGuid)
		// oss 中文件位置
		zipInfoTxt, err := oss.GetObjReader("yjreport", zipFile)
		if err != nil {
			log.Errorf("[fx] 获取zip-file.txt 失败%s, %s", zipFile, err)
			return ""
		}
		defer zipInfoTxt.Close()

		// 创建一个 Scanner 来逐行读取文件
		scanner := bufio.NewScanner(zipInfoTxt)
		if scanner.Scan() {
			fx.zipFileName = scanner.Text()
			return fx.zipFileName
		}
		return ""
	}
	return fx.zipFileName
}

func NewFxZipFileInfo(examGuid, templateGuid string) (zipFileInfo *FxZipFileInfo) {
	return &FxZipFileInfo{
		examGuid:     examGuid,
		templateGuid: templateGuid,
	}
}
