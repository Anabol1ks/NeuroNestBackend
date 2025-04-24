package router

import (
	"github.com/gin-gonic/gin"
)

func RouterConfig() *gin.Engine {
	r := gin.Default()
	return r
}
