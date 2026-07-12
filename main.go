package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"api-cliente/internal/auth"
	"api-cliente/internal/clientes"
	"api-cliente/internal/config"
	"api-cliente/internal/database"
	"api-cliente/internal/httpapi"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("configuracao invalida", "error", err)
		os.Exit(1)
	}

	db, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("erro ao conectar no banco", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	tokenManager := auth.NewTokenManager(cfg.JWTSecret, cfg.AccessTokenTTL)

	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo, tokenManager, cfg.RefreshTokenTTL, cfg.SessionTTL)
	authHandler := auth.NewHandler(authService)

	clienteRepo := clientes.NewRepository(db)
	clienteService := clientes.NewService(clienteRepo)
	clienteHandler := clientes.NewHandler(clienteService)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           httpapi.NewRouter(authHandler, clienteHandler, tokenManager),
		ReadHeaderTimeout: 5 * time.Second,
	}

	slog.Info("api iniciada", "addr", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("erro no servidor", "error", err)
		os.Exit(1)
	}
}
