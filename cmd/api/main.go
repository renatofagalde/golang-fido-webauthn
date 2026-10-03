package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/renatofagalde/golang-fido-webauthn/internal/fido/application"
	"github.com/renatofagalde/golang-fido-webauthn/internal/fido/infrastructure/repository"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	userRepository := repository.NewInMemoryUserRepository()
	userService := application.NewUserService(userRepository)
	fmt.Println(userService)

	slog.Info("server stating", "addr", "8080")
	// if err := router.Run(":8080"); err != nil {
	// 	slog.Error("server stopped", "error", err)
	// 	os.Exit(1)
	// }
}
