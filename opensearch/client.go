package opensearch

import (
	"github.com/SkyAPM/go2sky"
	"github.com/denverdino/aliyungo/common"
	"github.com/tiantianjianbao/craftsman/log"
)

const (
	Internet   = ""
	Intranet   = "intranet."
	VPC        = "vpc."
	APIVersion = "v2"
)

type OpenSearchConfig struct {
	IndexName       string `toml:"index_name"`
	NetWorkType     string `toml:"network_type"`
	Region          string `toml:"region"`
	AccessKeyId     string `toml:"accesskey_id"`
	AccessKeySecret string `toml:"accesskey_secret"`
	ApiVersion      string `toml:"api_version"`
}

type Client struct {
	common.Client
	indexName string
	tracer    *go2sky.Tracer
	logger    log.Logger
}

//OpenSearch的API比较奇怪，action不在公共参数里面
type OpenSearchArgs struct {
	Action string `ArgName:"action"`
}

func NewClient(config OpenSearchConfig, tracer *go2sky.Tracer, logger log.Logger) *Client {
	client := new(Client)
	client.Init("http://"+config.NetWorkType+"opensearch-"+config.Region+".aliyuncs.com", config.ApiVersion, config.AccessKeyId, config.AccessKeySecret)
	client.tracer = tracer
	client.logger = logger
	client.indexName = config.IndexName
	return client
}
