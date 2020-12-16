package transport

// 调度服务
type ServerClient struct {
	ServiceName     string `toml:"service_name"`
	Ipport          string `toml:"endpoints"`
	ProtoType       string `toml:"proto"`
	Balancetype     string `toml:"balancetype"`
	ConnectTimeout  int    `toml:"connnect_timeout"`
	ReadTimeout     int    `toml:"read_timeout"`
	WriteTimeout    int    `toml:"write_timeout"`
	MaxIdleConns    int    `toml:"max_idleconn"`
	RetryTimes      int    `toml:"retry_times"`
	SlowTime        int    `toml:"slow_time"`
	EndpointsFrom   string `toml:"endpoints_from"`
	LoadBalanceStat bool   `toml:"loadbalance_stat"`
}

//  提供服务
type Server struct {
	ServiceName string `toml:"service_name"` // 服务名称
	Port        int    `toml:"port"`         // 服务端口
	Proto       string `toml:"proto"`        // 服务协议
	ClusterName string `toml:"cluster_name"`
	GroupName   string `toml:"group_name"`
	RunModel    string `toml:"run_model"`
	TCP         struct {
		IdleTimeout      int `toml:"idle_timeout"`
		KeepliveInterval int `toml:"keeplive_interval"`
	} `toml:"tcp"`
}
