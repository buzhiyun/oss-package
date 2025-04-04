package oss

import (
	"context"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"github.com/buzhiyun/go-utils/cfg"
	"github.com/buzhiyun/go-utils/log"
)

var ossCfg = getOssCfg()

/**
 * 获取oss客户端
 */
func ossClient(client ...*oss.Client) *oss.Client {
	if len(client) > 0 {
		return client[0]
	}
	return oss.NewClient(ossCfg)
}

/**
 * 获取oss配置
 */
func getOssCfg() *oss.Config {
	key, ok := cfg.Config().GetString("oss.accessKeyId")
	if !ok {
		log.Error("获取oss.accessKeyId失败")
		return nil
	}
	secret, ok := cfg.Config().GetString("oss.accessKeySecret")
	if !ok {
		log.Error("获取oss.accessKeySecret失败")
		return nil
	}
	region, ok := cfg.Config().GetString("oss.region")
	if !ok {
		log.Error("获取oss.region失败")
		return nil
	}

	endpoint, ok := cfg.Config().GetString("oss.endpoint")
	if !ok {
		log.Error("获取oss.endpoint失败")
		return nil
	}

	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.CredentialsProviderFunc(func(ctx context.Context) (credentials.Credentials, error) {
			// 返回长期凭证
			return credentials.Credentials{AccessKeyID: key, AccessKeySecret: secret}, nil
			// 返回临时凭证
			//return credentials.Credentials{AccessKeyID: "id", AccessKeySecret: "secret",    SecurityToken: "token"}, nil
		})).
		WithRegion(region).WithEndpoint(endpoint)
	// WithEndpoint(endpoint)

	// 创建OSS客户端
	return cfg
}
