package fx

import (
	"bufio"
	"fmt"

	"strings"

	"github.com/buzhiyun/go-utils/cfg"
	"github.com/buzhiyun/go-utils/http"
	"github.com/buzhiyun/go-utils/log"
	"github.com/buzhiyun/oss-package/pkg/oss"
	"github.com/buzhiyun/oss-package/pkg/zip"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type objectInfo struct {
	bucketName string
	objectKey  string
	zipDir     string
}

type FxZipIndexFile struct {
	zip.GetZipfileInfo
	// zipInfoTxt   string
	examGuid       string
	templateGuid   string
	zipFileName    string
	zipInfoFileKey string // zipInfo文件在oss 的路径，从郑波接口上去拿
	// fileReadchan     chan *objectInfo // 给获取文件信息的线程缓冲读取oss文件的基础信息的
	// readWg           sync.WaitGroup
	fileList  *[]objectInfo
	fileCount int
	// fetchThreadCount int
	// bar              *uiprogress.Bar
}

func (fx *FxZipIndexFile) ListFileInfo(handle func(fileInfo *zip.SrcFileProperties)) {

	for _, obj := range *fx.fileList {
		_filename := strings.Split(obj.objectKey, "/")
		filename := fmt.Sprintf("%s/%s", obj.zipDir, _filename[len(_filename)-1])
		handle(&zip.SrcFileProperties{
			ZipDir:  &obj.zipDir,
			ZipPath: &filename,
			ObjKey:  &obj.objectKey,
		})
	}
	// // 启动多个线程去获取文件信息
	// for i := 0; i < 20; i++ {
	// 	fx.readWg.Add(1)
	// 	go func(id int, handle func(fileInfo *zip.SrcOssFile)) {
	// 		dl := oss.NewObjDownloader(i)
	// 		for {
	// 			select {
	// 			case obj, ok := <-fx.fileReadchan:
	// 				if !ok {
	// 					fx.readWg.Done()
	// 					log.Infof("[fx] 获取metadata线程-%v 完成", id)
	// 					return
	// 				}
	// 				if fx.bar != nil {
	// 					fx.bar.Set(len(fx.fileReadchan))
	// 				}

	// 				// 获取文件meta
	// 				var ossFileKey string
	// 				ossFileKey = obj.objectKey
	// 				result, err := dl.GetObjMeta(obj.bucketName, ossFileKey)
	// 				if err != nil {
	// 					log.Errorf("[fx] 获取oss meta 失败 %s, %s", ossFileKey, err)
	// 					continue
	// 				}
	// 				// log.Debugf("srcKey: %s, lastModified: %s, contentLength: %d", srcKey, *result.LastModified, result.ContentLength)
	// 				_filename := strings.Split(ossFileKey, "/")
	// 				filename := fmt.Sprintf("%s/%s", obj.zipDir, _filename[len(_filename)-1])
	// 				handle(&zip.SrcOssFile{
	// 					FileInfo: zip.NewZipFileInfo(false, filename, result.ContentLength, *result.LastModified, fs.ModePerm),
	// 					ObjKey:   &ossFileKey,
	// 					ZipDir:   &obj.zipDir,
	// 				})
	// 			}
	// 		}
	// 	}(i, handle)

	// 	// log.Debugf()
	// }

	// for _, obj := range *fx.fileList {
	// 	fx.fileReadchan <- &obj
	// }

	// // 当读取完成后
	// // 关闭文件读取通道
	// close(fx.fileReadchan)
	// // 等待所有线程完成
	// fx.readWg.Wait()

}

type zipPackageResponse struct {
	Status  int64  `json:"status"`
	Message string `json:"message"`
	Data    *Data  `json:"data,omitempty"`
}

type Data struct {
	OSSKey string `json:"OssKey"`
}

func (fx *FxZipIndexFile) fetchZipFile(bucket string) {
	// oss 中文件位置
	zipInfoTxt, err := oss.GetObjReader(bucket, fx.zipInfoFileKey)
	if err != nil {
		log.Fatalf("[fx] 获取zip-file.txt 失败%s, %s", zipInfoTxt, err)
		return
	}
	defer zipInfoTxt.Close()

	// 创建一个 Scanner 来逐行读取文件
	scanner := bufio.NewScanner(zipInfoTxt)
	var fileList []objectInfo
	lineNumber := 0
	// 逐行读取文件内容
	firstLine := true
	for scanner.Scan() {
		lineNumber++
		// log.Infof("第 %d 行: %s", lineNumber, scanner.Text())
		// 处理文件每一行
		if firstLine {
			fx.zipFileName = scanner.Text()
			log.Infof("[fx] zip文件 %s", fx.zipFileName)
			firstLine = false
			continue
		} else {
			line := strings.SplitN(scanner.Text(), "|", 2)
			if len(strings.TrimSpace(scanner.Text())) == 0 {
				continue
			}
			if len(line) < 2 {
				log.Errorf("[fx] 分隔符异常 第 %d 行: %%s", lineNumber, scanner.Text())
				continue
			}
			zipdir, ossKey := line[0], line[1]

			_objKey := strings.SplitN(ossKey, "/", 2)
			if len(_objKey) != 2 {
				log.Errorf("[fx] oss文件异常 第 %d 行: %s", lineNumber, scanner.Text())
				continue
			}

			srcBucket, srcKey := _objKey[0], _objKey[1]

			fileList = append(fileList, objectInfo{
				bucketName: srcBucket,
				objectKey:  srcKey,
				zipDir:     zipdir,
			})
		}
	}
	// 检查是否在读取过程中发生错误
	if err := scanner.Err(); err != nil {
		log.Infof("读取文件时发生错误:", err)
	}

	fx.fileList = &fileList
	fx.fileCount = len(fileList)

}

/*
*
  - 从分析中心接口去拿xing的zip文件
*/
func (fx *FxZipIndexFile) initZipFromFxService() {

	fxServiceApi, ok := cfg.Config().GetString("fx.api")
	if !ok {
		log.Fatal("获取fx.api 配置失败")
		return
	}

	packageApi := fmt.Sprintf("%s/FX.Service/BatchDownLoadDetail/PackageZip_zhouyang", fxServiceApi)

	document, _ := json.MarshalToString(map[string]string{
		"ExamGuid": fx.examGuid,
		"JobGuid":  fx.templateGuid,
	})
	resp, err := http.HttpPostForm(packageApi, map[string]string{
		"_septnet_document": document,
	}, http.HttpClientOption{
		Timeout: 0, // 加长超时
	})
	if err != nil {
		log.Fatalf("[fx] 获取zip文件信息失败 %s", err)
		return
	}
	log.Debugf("[fx] fx.service 接口返回: %s", resp)
	var zipPackageResp zipPackageResponse
	err = json.Unmarshal(resp, &zipPackageResp)
	if err != nil {
		log.Fatalf("[fx] 解析zip文件信息失败 %s ,\n%s", resp, err)
		return
	}

	// 解析 zipinfo 文件的bucket 和 key
	if zipPackageResp.Data == nil {
		log.Fatalf("[fx] fx.service 接口未拿到zip文件信息， %s", resp)
		return
	}
	zipListfileBucket, zipListfileKey, ok := oss.GetOssKeyFromLongKey(zipPackageResp.Data.OSSKey)
	if !ok {
		log.Fatalf("[fx] zip文件信息不对劲 %s", zipPackageResp.Data.OSSKey)
		return
	}

	fx.zipInfoFileKey = zipListfileKey

	fx.fetchZipFile(zipListfileBucket)

}

func NewFxZipFileInfoFromExam(examGuid, templateGuid string) (zipFileInfo *FxZipIndexFile) {
	_fzi := &FxZipIndexFile{
		examGuid:     examGuid,
		templateGuid: templateGuid,
		// fetchThreadCount: fetchThreadCount,
		// fileReadchan:     make(chan *objectInfo, fetchThreadCount*2),
	}
	log.Infof("准备开始获取 examGuid: %s, template: %s 的数据", examGuid, templateGuid)
	_fzi.initZipFromFxService()
	return _fzi
}

func NexFxZipFileInfoFromZipKey(zipLongkey string) (zipFileInfo *FxZipIndexFile) {
	bucket, key, ok := oss.GetOssKeyFromLongKey(zipLongkey)
	if !ok {
		log.Fatalf("[fx] zip文件key不正常 %s", zipLongkey)
		return nil
	}

	log.Infof("准备开始获取 %s 中的数据", zipLongkey)
	_fzi := &FxZipIndexFile{
		// fetchThreadCount: fetchThreadCount,
		zipInfoFileKey: key,
		// fileReadchan:     make(chan *objectInfo, fetchThreadCount*2),
	}
	_fzi.fetchZipFile(bucket)
	return _fzi
}

func (fx *FxZipIndexFile) GetZipFileName() string {
	return fx.zipFileName
}

func (fx *FxZipIndexFile) GetFileCount() int {
	return fx.fileCount
}

// func (fx *FxZipIndexFile) EnableChanProgressBar() {
// 	fx.bar = progress.NewProgressBar(fx.fetchThreadCount*2, "fetchmeta chan")
// }
