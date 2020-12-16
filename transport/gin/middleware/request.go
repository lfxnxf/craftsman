package middleware

import (
	"github.com/SkyAPM/go2sky"
	"github.com/garyburd/redigo/redis"
	"github.com/lfxnxf/craftsman/log"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lfxnxf/craftsman/transport"
)

const (
	ContextFieldBodyRaw    = "bodyRaw"
	ContextFieldDeviceInfo = "deviceInfo"
)

func Load(r *gin.Engine, log log.Logger, tracer *go2sky.Tracer, serviceName string) *gin.Engine {
	r.Use(trace(tracer, log), initRequest(log), logger(log, serviceName), recovery(log))
	return r
}

func initRequest(logger log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {

		//加载body信息
		bodyRaw := []byte("")
		if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut {
			reader := c.Request.Body
			defer reader.Close()
			b, err := ioutil.ReadAll(reader)
			if err == nil {
				bodyRaw = b
			}
		}

		c.Set(ContextFieldBodyRaw, bodyRaw)

		//加载device信息
		deviceInfo, _ := transport.NewDeviceWithRequest(c.Request, c.ClientIP())
		c.Set(ContextFieldDeviceInfo, deviceInfo)

		logger.DebugT(c.Request.Context(), "device info", redis.Args{}.AddFlat(deviceInfo)...)

		c.Next()
	}
}
