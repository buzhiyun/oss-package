package main

import (
	"flag"
	"fmt"
	"regexp"
	"runtime"

	"github.com/buzhiyun/go-utils/log"
	"github.com/buzhiyun/oss-package/internal/fx"
	"github.com/buzhiyun/oss-package/pkg/zip"
)

func main() {
	// prefix := flag.String("prefix", "excel/test/111", "要打包的oss路径")
	// zipFile := flag.String("zipkey", "excel/test/zip/111.zip", "打包后的zip文件路径")
	exam := flag.String("examguid", "", "考试guid")
	template := flag.String("template", "", "template guid")
	fxZipIndexFile := flag.String("file", "", "分析中心那边zip索引文件的路径 格式类似: yjreport/arithmeticcenterNew/TempFile/tmp/20250220-0805-363e-1e0b-a06e4b386593/ac73818d-39c9-40a1-9aa5-c313ebac1a1c/zip-path-file.txt")
	downloadThreadCount := flag.Int("dt", 10, "下载线程数")
	uploadThreadCount := flag.Int("ut", 3, "上传线程数")
	job := flag.String("job", "", "job名称 类似 fx_download_all_342025034_20250331-0803-2869-2fe1-afaa5b952482_4971e00d-6d65-4f6f-8035-12db9037d0f7_ffe254de-44bd-4ded-add5-081ca8bd3c56_576F4CC622DF3EDF5C66558E6B693D90")
	debug := flag.Bool("debug", false, "debug日志")
	flag.Parse()

	// windows 关闭日志颜色
	if runtime.GOOS == "windows" {
		log.DisableColor()
	}

	if *debug {
		log.SetLevel("debug")
	}

	var fxFile *fx.FxZipIndexFile
	if *fxZipIndexFile != "" {
		fxFile = fx.NexFxZipFileInfoFromZipKey(*fxZipIndexFile)
	} else if len(*job) > 0 {
		examguid, jobguid, ok := getExamFromJob(*job)
		if !ok {
			log.Fatalf("job名不正确 %s", *job)
			return
		}
		fxFile = fx.NewFxZipFileInfoFromExam(examguid, jobguid)
	} else if len(*exam) > 0 && len(*template) > 0 {
		fxFile = fx.NewFxZipFileInfoFromExam(*exam, *template)
	} else {
		log.Fatal("【 job | file | examguid templateguid 】 必须输入期中一组")
		return
	}

	// fxFile := fx.NewFxZipFileInfo(*exam, *template)

	z, err := zip.NewZipOssToOss("yjreport", fmt.Sprintf("yjreport/temp/%s", fxFile.GetZipFileName()), *downloadThreadCount, *uploadThreadCount, fxFile)
	if err != nil {
		panic(err)
	}
	z.Zip()
}

func getExamFromJob(jobName string) (examguid string, template string, ok bool) {
	// 定义正则表达式
	uuidPattern := `fx_download_(\w+)_(\d{7,9})_(\d{8}-\w{4}-\w{4}-\w{4}-\w{12})_(\w{8}-\w{4}-\w{4}-\w{4}-\w{12})_(\w{8}-\w{4}-\w{4}-\w{4}-\w{12})_(\w+)`
	// uuidPattern := `[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`
	ok = false
	re, err := regexp.Compile(uuidPattern)
	if err != nil {
		log.Errorf("正则表达式编译出错:", err)
		return
	}
	// 查找所有匹配的 UUID
	matches := re.FindStringSubmatch(jobName)
	log.Debug("jobName matches: %v", matches)
	if len(matches) >= 6 {
		return matches[3], matches[5], true
	}
	return
}
