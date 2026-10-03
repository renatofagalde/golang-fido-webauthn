package main

import (
	"log/slog"
	"os"

	"github.com/renatofagalde/golang-fido-webauthn/internal/fido/application"
	fidohttp "github.com/renatofagalde/golang-fido-webauthn/internal/fido/infrastructure/http"
	"github.com/renatofagalde/golang-fido-webauthn/internal/fido/infrastructure/repository"
	"github.com/renatofagalde/golang-fido-webauthn/internal/infrastructure/http"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	userRepository := repository.NewInMemoryUserRepository()
	userService := application.NewUserService(userRepository)
	userHandler := fidohttp.NewUserHandler(userService)

	router := http.NewRouter(http.RouterConfig{UserHandler: userHandler})

	slog.Info("server stating", "addr", "8080")
	if err := router.Run(":8080"); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
