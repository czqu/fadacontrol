package router

import "github.com/gin-gonic/gin"

type FadaControlRouter interface {
	Register()
	GetRouter() *gin.Engine
}
