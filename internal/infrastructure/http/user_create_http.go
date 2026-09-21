package http

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *userHandler) Create(c gin.Context) {
	var request CreateUserRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		slog.Warn("invalid request body",
			"flow_id", infrahttp.FlowIDFromContext(c.Request.Context()),
			"error", err,
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
}
