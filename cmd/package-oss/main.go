package main

import (
	"flag"

	"github.com/buzhiyun/go-utils/log"
	"github.com/buzhiyun/oss-package/pkg/zip"
)

func main() {
	prefix := flag.String("prefix", "yjreport/test/111", "要打包的oss路径")
	zipFile := flag.String("zipkey", "yjreport/test/zip/111.zip", "打包后的zip文件路径")
	downloadThreadCount := flag.Int("dt", 10, "下载线程数")
	uploadThreadCount := flag.Int("ut", 3, "上传线程数")
	debug := flag.Bool("debug", false, "debug日志")
	flag.Parse()

	if *debug {
		log.SetLevel("debug")
	}

	z, err := zip.NewZipOssToOss(*prefix, *zipFile, *downloadThreadCount, *uploadThreadCount)
	if err != nil {
		panic(err)
	}
	z.Zip()
}
