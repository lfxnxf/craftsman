package sd

// 服务发现
type ServiceDiscovery struct {
	ServiceName   string `toml:"service_name"` //
	NamespaceId   string `toml:"namespace_id"`
	Clusters      string `toml:"clusters"`
	Ipport        string `toml:"endpoints"`      //
	Balancetype   string `toml:"balancetype"`    // 负载均衡类型
	ProtoType     string `toml:"proto"`          // 服务协议
	EndpointsFrom string `toml:"endpoints_from"` // 节点来源
}
