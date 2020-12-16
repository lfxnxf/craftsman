package nacos

import (
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/model"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

type Client struct {
	nacos naming_client.INamingClient
}

func NewNamingClient(nc naming_client.INamingClient) (Client, error) {
	return Client{nacos: nc}, nil
}

func (c *Client) RegisterInstance(param vo.RegisterInstanceParam) (bool, error) {
	return c.nacos.RegisterInstance(param)
}

func (c *Client) DeregisterInstance(param vo.DeregisterInstanceParam) (bool, error) {
	return c.nacos.DeregisterInstance(param)
}

func (c *Client) GetService(param vo.GetServiceParam) (model.Service, error) {
	return c.nacos.GetService(param)
}

func (c *Client) GetAllServicesInfo(param vo.GetAllServiceInfoParam) ([]model.Service, error) {
	return c.nacos.GetAllServicesInfo(param)
}

func (c *Client) SelectAllInstances(param vo.SelectAllInstancesParam) ([]model.Instance, error) {
	return c.nacos.SelectAllInstances(param)
}

func (c *Client) SelectInstances(param vo.SelectInstancesParam) ([]model.Instance, error) {
	return c.nacos.SelectInstances(param)
}

func (c *Client) SelectOneHealthyInstance(param vo.SelectOneHealthInstanceParam) (*model.Instance, error) {
	return c.nacos.SelectOneHealthyInstance(param)
}

func (c *Client) Subscribe(param *vo.SubscribeParam) error {
	return c.nacos.Subscribe(param)
}

func (c *Client) Unsubscribe(param *vo.SubscribeParam) error {
	return c.nacos.Unsubscribe(param)
}
