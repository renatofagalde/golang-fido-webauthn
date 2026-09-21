package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/renatofagalde/golang-fido-webauthn/internal/fido/domain"
)

func writeDomainError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		writeerror.JSON(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrUsernametaken):
		writeerror.JSON(c, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrUserNotFound):
		writeerror.JSON(c, http.StatusNotFound, err.Error())
	default:
		writeerror.JSON(c, http.StatusInternalServerError, "internal error")
	}
}
