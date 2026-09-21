package http

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/renatofagalde/golang-fido-webauthn/pkg/flowid"
)

func (h *userHandler) Create(c gin.Context) {
	var request CreateUserRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		slog.Warn("invalid request body",
			"flow_id", flowid.FromContext(c.Request.Context()),
			"error", err,
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	user, err := h.service.Create(c.Request.Context(), toCreateUserInput(request))
	if err != nil {
		writeDomainError(c, err)
		return
	}

	fmt.Println(user)
}
