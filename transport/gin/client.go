package gin

import (
	"github.com/SkyAPM/go2sky"
	"github.com/gin-gonic/gin"
	"github.com/tiantianjianbao/craftsman/log"
	"github.com/tiantianjianbao/craftsman/transport/gin/middleware"
)

type Client struct {
	R      *gin.Engine
	Resp   Response
	Logger log.Logger
}

func NewClient(serviceName string, runModel string, logger log.Logger, tracer *go2sky.Tracer) *Client {
	gin.SetMode(runModel)
	r := gin.New()
	middleware.Load(r, logger, tracer, serviceName)

	return &Client{
		R:      r,
		Logger: logger,
		Resp:   NewResponse(logger),
	}
}
