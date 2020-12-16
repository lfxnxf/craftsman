package inits

import (
	"strconv"
	"strings"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/model"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/tiantianjianbao/craftsman/log"
	"github.com/tiantianjianbao/craftsman/sd"
	"github.com/tiantianjianbao/craftsman/sd/nacos"
)

var (
	GlobalNacos *NacosClient
)

type NacosClient struct {
	NacosClient nacos.Client
	logger      log.Logger
	BalanceType string
}

func NewNacosClient(log log.Logger, config sd.ServiceDiscovery) (*NacosClient, error) {
	nacosClient := &NacosClient{
		logger:      log,
		BalanceType: config.Balancetype,
	}

	args := strings.Split(config.Ipport, ":")
	port, err := strconv.Atoi(args[1])
	if err != nil {
		log.Error("ipport", "err", err)
		return nacosClient, err
	}

	namingClient, err := clients.CreateNamingClient(map[string]interface{}{
		"serverConfigs": []constant.ServerConfig{
			{
				IpAddr:      args[0],
				Port:        uint64(port),
				ContextPath: "/nacos", //nacos服务的上下文路径，默认是“/nacos”
			},
		},
		"clientConfig": constant.ClientConfig{
			TimeoutMs:            10 * 100,            //http请求超时时间，单位毫秒
			ListenInterval:       30 * 100,            //监听间隔时间，单位毫秒（仅在ConfigClient中有效）
			BeatInterval:         5 * 100,             //心跳间隔时间，单位毫秒（仅在ServiceClient中有效）
			NamespaceId:          config.NamespaceId,  //nacos命名空间
			Endpoint:             "",                  //获取nacos节点ip的服务地址
			CacheDir:             "/data/nacos/cache", //缓存目录
			UpdateThreadNum:      20,                  //更新服务的线程数
			NotLoadCacheAtStart:  true,                //在启动时不读取本地缓存数据，true--不读取，false--读取
			UpdateCacheWhenEmpty: true,                //当服务列表为空时是否更新本地缓存，true--更新,false--不更新
		},
	})
	if err != nil {
		log.Error("create client", "err", err)
		return nacosClient, err
	}

	client, err := nacos.NewNamingClient(namingClient)
	if err != nil {
		log.Error("create client", "err", err)
		return nacosClient, err
	}

	nacosClient.NacosClient = client

	GlobalNacos = nacosClient

	return nacosClient, nil
}

type HealthyInstance func() (*model.Instance, error)

func GetOneHealthyInstance(service string) HealthyInstance {
	return func() (*model.Instance, error) {
		return GlobalNacos.NacosClient.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
			ServiceName: service,
			Clusters:    []string{},
		})
	}
}

func GetAllHealthyInstance(service string) ([]model.Instance, error) {
	return GlobalNacos.NacosClient.SelectAllInstances(vo.SelectAllInstancesParam{
		ServiceName: service,
	})
}
