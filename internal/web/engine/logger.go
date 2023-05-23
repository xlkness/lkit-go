package engine

import (
	"time"

	"github.com/gin-gonic/gin"
)

type Logger interface {
	HandleRequest(ctx Context, params *gin.LogFormatterParams)
}

type defaultLogger struct {
}

func (l *defaultLogger) HandleRequest(ctx Context, params *gin.LogFormatterParams) {

}

func parseContextLogParams(ctx Context) *gin.LogFormatterParams {
	start := time.Now()
	c := ctx.GetGinContext()
	path := c.Request.URL.Path
	raw := c.Request.URL.RawQuery

	// Process request
	c.Next()

	// Log only when path is not being skipped

	param := &gin.LogFormatterParams{
		Request: c.Request,
		Keys:    c.Keys,
	}

	// Stop timer
	param.TimeStamp = time.Now()
	param.Latency = param.TimeStamp.Sub(start)

	param.ClientIP = c.ClientIP()
	param.Method = c.Request.Method
	param.StatusCode = c.Writer.Status()
	param.ErrorMessage = c.Errors.ByType(gin.ErrorTypePrivate).String()

	param.BodySize = c.Writer.Size()

	if raw != "" {
		path = path + "?" + raw
	}

	param.Path = path

	return param
}
