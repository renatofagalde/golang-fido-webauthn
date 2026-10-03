package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	userHTTP "github.com/renatofagalde/golang-fido-webauthn/internal/fido/infrastructure/http"
)

type RouterConfig struct {
	userHTTP.UserHandler
	userHandler userHTTP.UserHandler
}

func NewRouter(cfg RouterConfig) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/handle", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/user", FlowID())

	fido := api.Group("/fido")
	users := fido.Group("/users")
	{
		users.POST("", cfg.UserHandler.Create)
	}
	return r
}
