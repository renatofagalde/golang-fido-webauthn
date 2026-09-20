package http

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// um tipo proprio para chaves
type contextKey string

const flowIDKey contextKey = "flow_id"

const headerFlowID = "X-Flow-ID"

// FLowID	é o middleware mais externo da stack
// O mesmo id por requisicao, propaga para as camadas abaixo
// logarem com o mesmo FLowID

func FlowID() gin.HandlerFunc {
	return func(c *gin.Context) {
		flowID := c.GetHeader(headerFlowID)

		// headerFlowID obrigatorio
		if _, err := uuid.Parse(flowID); err != nil {
			slog.Error("invalid X-Flow-ID",
				"method", c.Request.Method,
				"path", c.Request.URL.Path)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid X-Flow-ID"})
			return
		}
	}
}
