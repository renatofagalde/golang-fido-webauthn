package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	infraHTTP "github.com/renatofagalde/golang-fido-webauthn/internal/infrastructure/http"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(infraHTTP.FlowID())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	slog.Info("server stating", "addr", "8080")
	if err := router.Run(":8080"); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
