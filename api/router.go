package api

import (
	"github.com/gin-gonic/gin"
)

var (
	router *gin.Engine
)

func Router() *gin.Engine {
	if router == nil {
		gin.SetMode(gin.ReleaseMode)

		router = gin.New()
		router.Use(gin.Logger(), FaultWrap())
	}
	return router
}
