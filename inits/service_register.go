package inits

import (
	"errors"
	"net"

	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/lfxnxf/craftsman/log"
)

func RegisterInstance(log log.Logger, config Config, nc *NacosClient) error {
	localIp := LocalIPString(log)
	if localIp == "" {
		return errors.New("ip err")
	}

	serviceName := config.GetServiceName()
	if serviceName == "" {
		return errors.New("service name null")
	}

	port := config.GetServicePort()
	if port < 1 {
		return errors.New("port null")
	}

	res, err := nc.NacosClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          localIp,
		Port:        uint64(port),
		ServiceName: serviceName,
		Weight:      10,
		ClusterName: config.GetServiceClusterName(),
		GroupName:   config.GetServiceGroupName(),
		Enable:      true,
		Healthy:     false,
		Ephemeral:   true,
	})
	if err != nil {
		log.Error("nacos register", "err:", err)
		return nil
	}
	log.Info("nacos register", "register", res,
		"localIp", localIp, "serviceName", serviceName, "port", port)
	return nil
}

// LocalIP tries to determine a non-loopback address for the local machine
func LocalIP() (net.IP, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.IsGlobalUnicast() {
			if ipnet.IP.To4() != nil || ipnet.IP.To16() != nil {
				return ipnet.IP, nil
			}
		}
	}
	return nil, nil
}

func LocalIPString(log log.Logger) string {
	ip, err := LocalIP()
	if err != nil {
		log.Error("error determining local ip address. ", err)
		return ""
	}
	if ip == nil {
		log.Error("could not determine local ip address")
		return ""
	}
	return ip.String()
}
