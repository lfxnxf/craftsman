package middleware

import (
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
	"github.com/lfxnxf/craftsman/log"
	"github.com/lfxnxf/craftsman/transport"
	"net/http"
	"time"
)

func logger(logger log.Logger, serviceName string) gin.HandlerFunc {

	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		if raw != "" {
			path = path + "?" + raw
		}

		var data string
		if bodyRaw, ok := c.Get(ContextFieldBodyRaw); ok {
			if d, ok := bodyRaw.([]byte); ok {
				data = string(d)
			}
		}

		requestInfo := requestInfo{
			RequestUrl: path,
			Data:       data,
			Method:     c.Request.Method,
			StatusCode: c.Writer.Status(),
			BodySize:   c.Writer.Size(),
		}

		RequestInfoStr, _ := json.Marshal(requestInfo)
		latency := time.Now().Sub(start)

		logInfo := make(map[string]string, 3)
		logInfo["request_id"] = ""
		logInfo["client_ip"] = c.ClientIP()
		logInfo["latency"] = latency.String()
		logInfo["request_info"] = string(RequestInfoStr)
		logInfo["service_name"] = c.GetString(serviceName)
		d, ok := c.Get(ContextFieldDeviceInfo)
		if ok {
			dInfo := d.(*transport.Device)
			logInfo["user_id"] = dInfo.UserId
			logInfo["gateway_ip"] = dInfo.GatewayIp
		}

		if c.Writer.Status() != http.StatusOK {
			logInfo["err"] = c.Errors.ByType(gin.ErrorTypePrivate).String()
			logger.ErrorT(c.Request.Context(), "request Err", redis.Args{}.AddFlat(logInfo)...)
		} else {
			logger.InfoT(c.Request.Context(), "request OK", redis.Args{}.AddFlat(logInfo)...)
		}
	}
}

type requestInfo struct {
	RequestUrl string `json:"request_url"`
	Method     string `json:"method"`
	Data       string `json:"data"`
	StatusCode int    `json:"status_code"`
	BodySize   int    `json:"body_size"`
}
