package transport

import (
	"github.com/spf13/cast"
	"net/http"
	"strconv"
	"strings"
)

const (
	HttpHeaderWidth     = "X-APP-Width"
	HttpHeaderHeight    = "X-APP-Height"
	HttpHeaderID        = "X-APP-Id"
	HttpHeaderVersion   = "X-APP-Version"
	HttpHeaderChannel   = "X-APP-Channel"
	HttpHeaderModel     = "X-APP-Model"
	HttpHeaderName      = "X-APP-Name"
	HttpHeaderIMEI      = "X-Device-IMEI"
	HttpHeaderGatewayIp = "X-GATEWAY-IP"
	HttpHeaderUserId    = "X-USER-ID"
)

type Device struct {
	Platform     string
	AppVersion   string
	UserId       string
	DeviceUniqId string
	FromChannel  string
	Model        string
	Brand        string
	ScreenW      int
	ScreenH      int
	GatewayIp    string
	ClientIp     string
}

func NewDeviceWithRequest(c *http.Request, clientIP string) (*Device, error) {
	deviceScreenW, _ := strconv.Atoi(c.Header.Get(HttpHeaderWidth))
	deviceScreenH, _ := strconv.Atoi(c.Header.Get(HttpHeaderHeight))

	deviceInfo := new(Device)
	deviceInfo.Platform = c.Header.Get(HttpHeaderID)
	deviceInfo.AppVersion = c.Header.Get(HttpHeaderVersion)
	deviceInfo.FromChannel = c.Header.Get(HttpHeaderChannel)
	deviceInfo.Model = c.Header.Get(HttpHeaderModel)
	deviceInfo.Brand = c.Header.Get(HttpHeaderName)
	deviceInfo.ScreenW = deviceScreenW
	deviceInfo.ScreenH = deviceScreenH
	deviceInfo.GatewayIp = c.Header.Get(HttpHeaderGatewayIp)
	deviceInfo.DeviceUniqId = strings.Trim(c.Header.Get(HttpHeaderIMEI), "")
	deviceInfo.UserId = cast.ToString(c.Header.Get(HttpHeaderUserId))
	deviceInfo.ClientIp = clientIP

	return deviceInfo, nil
}
