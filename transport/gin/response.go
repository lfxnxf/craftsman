package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/lfxnxf/craftsman/log"
	"github.com/lfxnxf/craftsman/transport/gin/middleware"
	"net/http"
)

const (
	CodeSuccess     int32  = 1000
	ContentTypeJson string = "json"
)

type Response interface {
	Success(ctx *gin.Context, d interface{})
	Error(ctx *gin.Context, code int32, message string)
	LoadMessage(messageMap map[int32]string)
}

func NewResponse(logger log.Logger) Response {
	return &response{
		logger:     logger,
		messageMap: map[int32]string{},
	}
}

type response struct {
	messageMap map[int32]string
	logger     log.Logger
}

type reply struct {
	Code    int32       `json:"code"`
	Data    interface{} `json:"data,string"`
	Message string      `json:"message"`
}

func (r *response) Success(ctx *gin.Context, d interface{}) {
	reply := &reply{
		Code:    CodeSuccess,
		Data:    d,
		Message: "",
	}

	r.sendData(ctx, reply, ContentTypeJson)
}

func (r *response) Error(ctx *gin.Context, code int32, message string) {
	reply := &reply{
		Code:    code,
		Data:    "",
		Message: message,
	}

	r.sendData(ctx, reply, ContentTypeJson)
}

func (r *response) LoadMessage(messageMap map[int32]string) {
	if messageMap != nil {
		for k, v := range messageMap {
			r.messageMap[k] = v
		}
	}
}

func (r *response) sendData(ctx *gin.Context, reply *reply, format string) {
	originMessage := reply.Message
	if msg, ok := r.messageMap[reply.Code]; ok {
		reply.Message = msg
	}

	switch format {
	case ContentTypeJson:
		ctx.JSON(http.StatusOK, reply)
	}
	ctx.Abort()

	if reply.Code != CodeSuccess {
		bodyStr := ""
		body, ok := ctx.Get(middleware.ContextFieldBodyRaw)
		if ok {
			bodyStr = string(body.([]byte))
		}
		r.logger.ErrorT(ctx.Request.Context(), "response err", "err", originMessage, "code", reply.Code, "url", ctx.Request.URL.String(), bodyStr)
	}
}
