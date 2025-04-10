package zip

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strings"
	"sync"

	"github.com/buzhiyun/oss-package/pkg/oss"
	"github.com/buzhiyun/oss-package/pkg/progress"

	"github.com/buzhiyun/go-utils/log"
)

type SrcOssFile struct {
	FileInfo fs.FileInfo // 给 zip 压缩header 用的
	ZipDir   *string     // 给 zip 压缩所在的目录
	ObjKey   *string     // oss文件key
	data     *[]byte
}

type zipOssToOss struct {
	srcBucketName string
	// srcPrefix           string
	zipBucketName       string
	zipFileKey          string
	zipfileInfo         GetZipfileInfo  //  获取zip文件列表实体
	downloadThreadCount int             // 下载线程数
	uploadThreadCount   int             // 上传线程数
	downloadChan        chan SrcOssFile // 下载队列    ，给下载线程用的，实际下载线程会用到 oss2.ObjectProperties 里面的 Key  Size  LastModified  三个属性
	zipChan             chan SrcOssFile // 压缩队列
	wg                  sync.WaitGroup
	zipWg               sync.WaitGroup
	options             *ZipOption
}

/**
 * 初始化oss打包实例
 */
func NewZipOssToOss(srcBucketName, zipFileKey string, downloadThreadCount, uploadThreadCount int, zipfileInfoInstance GetZipfileInfo) (z *zipOssToOss, err error) {
	_zippath := strings.SplitN(zipFileKey, "/", 2)

	if downloadThreadCount == 0 || uploadThreadCount == 0 {
		err = errors.New("[zip] 下载线程数或上传线程数不能为0")
		log.Fatal(err)
		return
	}

	if len(_zippath) == 2 {
		// srcBucketName := _path[0]
		// srcPrefix := _path[1]

		zipBucketName := _zippath[0]
		zipKey := _zippath[1]
		return &zipOssToOss{
			options:             &DefaultZipOption,
			srcBucketName:       srcBucketName,
			zipBucketName:       zipBucketName,
			zipFileKey:          zipKey,
			downloadThreadCount: downloadThreadCount,
			uploadThreadCount:   uploadThreadCount,
			zipfileInfo:         zipfileInfoInstance,
			downloadChan:        make(chan SrcOssFile, downloadThreadCount*2),
			zipChan:             make(chan SrcOssFile, downloadThreadCount*2),
		}, nil
	}
	err = errors.New("[zip] 压缩文件路径错误")
	log.Fatal(err)
	return
}

/**
 * 下载oss文件
 */
func (z *zipOssToOss) downloadOssObj(obj *SrcOssFile, dl *oss.ObjDownloader) {
	// log.Infof("[zip] 下载文件")
	data, err := dl.Download(z.srcBucketName, *obj.ObjKey)
	if err != nil {
		log.Errorf("[zip] downloadThread-%v 下载文件失败", dl.GetId())
		return
	}
	// 丢入压缩队列
	// fi := getFileInfo(obj)
	obj.data = &data
	z.zipChan <- *obj
}

func zipDirHandle(dirMap *map[string]any, dir string, handle func(dir string)) {
	if _, ok := (*dirMap)[dir]; ok {
		return
	}

	//递归查找上级目录
	_dir := strings.Split(dir, "/")
	if len(_dir) > 1 {
		parentDir := strings.Join(_dir[:len(_dir)-1], "/")
		// log.Debugf("dir: %s , parentDir: %s", dir, parentDir)
		zipDirHandle(dirMap, parentDir, handle)
	}

	handle(dir)
	(*dirMap)[dir] = 1
}

/**
 * 压缩文件
 */
func (z *zipOssToOss) Zip(zipOptions ...func(*ZipOption)) {
	for _, fn := range zipOptions {
		fn(z.options)
	}

	var useProgress = z.options.ProgressBar
	progressBar := progress.NewProgressBar(int64(z.options.TotalFileCount))
	if useProgress {
		log.SetLevel("error")
		fmt.Println("  下载中...")
		progress.EnableProgressBar()
	}

	zipfile, err := oss.NewMultipartUploadWriter(z.zipBucketName, z.zipFileKey, z.uploadThreadCount)
	if err != nil {
		return
	}
	defer zipfile.Close()
	// 打开：zip文件
	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	for i := range z.downloadThreadCount {
		log.Infof("[zip] downloadThread-%v 开始下载文件", i)
		// 启动下载线程方法
		z.wg.Add(1)
		go func(threadId int) {
			dl := oss.NewObjDownloader(threadId)
			for {
				select {
				case srcFile, ok := <-z.downloadChan:
					if !ok {
						z.wg.Done()
						log.Infof("[zip] downloadThread-%v 下载文件任务完成", i)
						return
					}
					z.downloadOssObj(&srcFile, dl)
					progressBar.Incr()
				}
			}
		}(i)
	}

	// 去拿oss目录下的文件列表
	z.wg.Add(1)
	go func() {
		// oss.ListPathWithHandle(z.srcBucketName, z.srcPrefix, func(obj oss2.ObjectProperties) {
		// 	log.Debugf("获取到 %s", *obj.Key)
		// 	z.downloadChan <- obj
		// })
		z.zipfileInfo.ListFileInfo(func(obj *SrcOssFile) {
			log.Debugf("获取到 %s", *obj.ObjKey)
			z.downloadChan <- *obj
		})
		z.wg.Done()
		// 关闭下载队列，让下载线程正常退出
		close(z.downloadChan)
	}()

	// 压缩线程
	log.Info("启动压缩线程")
	z.zipWg.Add(1)
	go func() {
		// var zipDirMap = make(map[string]any)
		// var currentDir string
		for {
			select {
			case srcFile, ok := <-z.zipChan:
				if !ok {
					log.Infof("[zip] 压缩文件任务完成")
					z.zipWg.Done()
					return
				}

				// currentDir = *srcFile.ZipDir

				log.Debugf("[zip] 压缩 %s %s", *srcFile.ZipDir, *srcFile.ObjKey)
				fi := srcFile.FileInfo

				// //检查创建目录
				// zipDirHandle(&zipDirMap, currentDir, func(dir string) {
				// 	// 可能有问题
				// 	dirInfo := NewZipFileInfo(true, dir, 1024, fi.ModTime(), fs.ModeDir|fs.ModePerm)
				// 	dirHeader, _ := zip.FileInfoHeader(dirInfo)
				// 	log.Infof("[zip] 创建目录 %s , %v", dirInfo.Name(), dirInfo.IsDir())
				// 	archive.CreateHeader(dirHeader)
				// })

				header, _ := zip.FileInfoHeader(fi)

				// 这地方要改
				header.Name = fi.Name()
				// header.Name = strings.TrimPrefix(*srcFile.objKey, z.srcPrefix)
				header.Name = strings.TrimPrefix(header.Name, "/")

				// 判断：文件是不是文件夹
				if !fi.IsDir() {
					// 设置：zip的文件压缩算法
					header.Method = zip.Deflate
				} else if header.Name == "" {
					break
				}

				// 创建：压缩包头部信息
				writer, err := archive.CreateHeader(header)
				log.Debugf("[zip] 创建 %s", header.Name)
				if err != nil {
					log.Errorf("[zip] 创建 %s 异常: %v", header.Name, err)
					break
				}
				if !fi.IsDir() {
					// 读取obj
					br := bytes.NewReader(*srcFile.data)
					io.Copy(writer, br)
				}
			}
		}
	}()

	// 等待下载线程退出
	z.wg.Wait()
	close(z.zipChan)

	// 等待压缩线程退出
	z.zipWg.Wait()
}
