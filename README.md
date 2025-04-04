## 阿里云OSS打包器

`由于阿里云OSS的文件不支持在线打包，这个工具可以快速将某个OSS目录下的资源打包到OSS上，方便后续下载。`

- 方面大量小文件快速打包

- 使用方法
  ```bash
  # 打包 oss://report/2025 下的所有文件
  # 打包后的zip文件为 oss://report/test/zip/111.zip，上传线程数为30，下载线程数为300
  ./package-oss -prefix report/2025 -zipkey report/test/zip/111.zip -ut 30 -dt 300

  Usage of ./bin/package-oss:
  -debug
        debug日志
  -dt int
        下载线程数 (default 10)
  -prefix string
        要打包的oss路径 (default "excel/test/111")
  -ut int
        上传线程数 (default 3)
  -zipkey string
        打包后的zip文件路径 (default "excel/test/zip/111.zip")
  ```

- 配置文件 config.yaml
  ```yaml
  oss:
    region: "cn-hangzhou"    # oss区域 必须配置
    # 由于阿里云有dns查询有限制，所以尽量把Bucket域名给绑  定在hosts文件上，否则并发高的下载容易触发ECS解析查询熔断
    # https://help.aliyun.com/zh/dns/how-can-the-speed-limit-of-ecs-dns-query-requests
    # endpoint: "http://oss-cn-hangzhou-internal.aliyuncs.com"    # 内网访问oss，可以用http
    endpoint: "oss-cn-hangzhou.aliyuncs.com"
    accessKeyId:   "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
    accessKeySecret:   "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  ```