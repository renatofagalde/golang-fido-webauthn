package writeerror

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/renatofagalde/golang-fido-webauthn/pkg/flowid"
)

func JSON(c *gin.Context, status int, message string) {
	if status >= http.StatusInternalServerError {
		slog.Error("request failed",
			"flow_id", flowid.FromContext(c.Request.Context()),
			"status", status,
			"message", message,
		)
	} else {
		slog.Warn("request rejected",
			"flow_id", flowid.FromContext(c.Request.Context()),
			"status", status,
			"message", message,
		)
	}
	c.JSON(status, gin.H{"error": message})
}
