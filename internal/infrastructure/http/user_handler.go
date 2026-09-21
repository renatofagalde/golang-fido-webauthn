package http

import (
	"github.com/gin-gonic/gin"
	"github.com/renatofagalde/golang-fido-webauthn/internal/fido/application"
)

type UserHandler interface {
	Create(c *gin.Context)
}

type userHandler struct {
	service application.UserUsecase
}

func NewUserHandler(service application.UserUsecase) UserHandler {
	return &userHandler{service: service}
}
