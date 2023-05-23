package engine

import (
	"github.com/gin-gonic/gin"
)

type Context interface {
	SetGinContext(ctx *gin.Context)
	GetGinContext() *gin.Context
	ResponseParseParamsFieldFail(path string, uri string, body string, field string, value string, err error)
}
