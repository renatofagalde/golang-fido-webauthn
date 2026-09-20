// Package http define rotas e middleware
package http

import (
	"context"
	"log/slog"
	"net/http"
	"time"

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

		ctx := context.WithValue(c.Request.Context(), flowIDKey, flowID)
		c.Request = c.Request.WithContext(ctx)
		c.Writer.Header().Set(headerFlowID, flowID)

		start := time.Now()

		slog.Info("request received", "flow_id", flowID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path)

		c.Next()

		slog.Info("request completed", "flow_id", flowID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency", time.Since(start).Milliseconds())
	}
}
