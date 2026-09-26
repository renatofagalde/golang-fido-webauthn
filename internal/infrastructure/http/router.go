package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	fidohttp "github.com/renatofagalde/golang-fido-webauthn/internal/infrasstructure/http"
)

type RouterConfig struct {
	UserHandler fidohttp.UserHandler
}

func NewRouter(cfg RouterConfig) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/handle", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return r
}
