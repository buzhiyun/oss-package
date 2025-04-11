## 阿里云OSS打包器

`由于阿里云OSS的文件不支持在线打包，这个工具可以快速将某个OSS目录下的资源打包到OSS上，方便后续下载。`

- 方面大量小文件快速打包

- 使用方法
  ```bash
  # 打包 fx_download_all_342025034_20250331-0803-2869-2fe1-afaa5b952482_4971e00d-6d65-4f6f-8035-12db9037d0f7_ffe254de-44bd-4ded-add5-081ca8bd3c56_576F4CC622DF3EDF5C66558E6B693D90 下的所有文件
  # 打包后的zip文件为 oss://yjreport/test/xxxxxxx.zip，上传线程数为10，下载线程数为300
  ./package-oss -job 'fx_download_all_342025034_20250331-0803-2869-2fe1-afaa5b952482_4971e00d-6d65-4f6f-8035-12db9037d0f7_ffe254de-44bd-4ded-add5-081ca8bd3c56_576F4CC622DF3EDF5C66558E6B693D90' -dt 300 -ut 10 -g

  Usage of ./bin/package-oss:
      -debug
            debug日志
      -dt int
            下载线程数 (default 256)
      -ut int
            上传线程数 (default 8)
      -examguid string
            考试guid
      -template string
            template guid
      -file string
            分析中心那边zip索引文件的路径 格式类似: yjreport/arithmeticcenterNew/    TempFile/tmp/20250220-0805-363e-1e0b-a06e4b386593/    ac73818d-39c9-40a1-9aa5-c313ebac1a1c/zip-path-file.txt
      -job string
            job名称 类似     fx_download_all_342025034_20250331-0803-2869-2fe1-afaa5b952482_4971e    00d-6d65-4f6f-8035-12db9037d0f7_ffe254de-44bd-4ded-add5-081ca8bd3c56    _576F4CC622DF3EDF5C66558E6B693D90
      -g bool
            是否显示进度条
      -zl int
            zip压缩级别 0-9 ， 越高压缩越慢效果越好 (default 1)
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

  fx:
    api: http://vfs.stg1:7011
  ```