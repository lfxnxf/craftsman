package tracing

// 链路服务
type Trace struct {
	ServiceName   string `toml:"service_name"`   // 链路服务名称
	Ipport        string `toml:"endpoints"`      // 链路服务节点
	Balancetype   string `toml:"balancetype"`    // 负载均衡类型
	ProtoType     string `toml:"proto"`          // 服务协议
	EndpointsFrom string `toml:"endpoints_from"` // 节点来源
}
