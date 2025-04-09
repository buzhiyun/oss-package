package main

import (
	"flag"
	"fmt"
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
	downloadThreadCount := flag.Int("dt", 10, "下载线程数")
	uploadThreadCount := flag.Int("ut", 3, "上传线程数")
	debug := flag.Bool("debug", false, "debug日志")
	flag.Parse()

	// windows 关闭日志颜色
	if runtime.GOOS == "windows" {
		log.DisableColor()
	}

	if *debug {
		log.SetLevel("debug")
	}

	fxFile := fx.NewFxZipFileInfo(*exam, *template)

	z, err := zip.NewZipOssToOss("yjreport", fmt.Sprintf("yjreport/test/%s", fxFile.GetZipFileName()), *downloadThreadCount, *uploadThreadCount, fxFile)
	if err != nil {
		panic(err)
	}
	z.Zip()
}
