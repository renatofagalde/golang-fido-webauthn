package main

import (
	"log/slog"
	"os"

	"github.com/renatofagalde/golang-fido-webauthn/internal/fido/application"
	"github.com/renatofagalde/golang-fido-webauthn/internal/fido/infrastructure/repository"
	fidoHTTP "github.com/renatofagalde/golang-fido-webauthn/internal/infrastructure/http"
	infraHTTP "github.com/renatofagalde/golang-fido-webauthn/internal/infrastructure/http"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	userRepository := repository.NewInMemoryUserRepository()
	userService := application.NewUserService(userRepository)
	userHandler := fidoHTTP.NewUserHandler(userService)

	infraHTTP.NewRouter

	slog.Info("server stating", "addr", "8080")
	if err := router.Run(":8080"); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
