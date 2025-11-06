package controller

import "github.com/gin-gonic/gin"

// Handler 基础处理器接口
type Handler interface {
	Register(router *gin.RouterGroup)
}
