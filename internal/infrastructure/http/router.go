package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)


type RouterConfig struct{
	UserHandler UserHandler
}
func NewRouter(config RouterConfig)*gin.Engine{
	
	r:=gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK,gin.H{"status":"ok"})
	})

	api:=r.Group("/"gin.HandlerFunc)


}


