package main

import (
"context"
"fmt"
"net/http"
"os"
"os/signal"
"syscall"
"time"

"github.com/abhishekmaurya/url-shortner/repo"
"github.com/abhishekmaurya/url-shortner/services"
"github.com/abhishekmaurya/url-shortner/services/urlshortener"
"github.com/abhishekmaurya/url-shortner/utils"
)

func main() {
	utils.InitLogger()

	cfg := utils.GetConfig()

	utils.Info("starting url-shortener service", "port", cfg.Server.Port)

	// Initialize repository (handles DB connection + migrations)
	repoManager, err := repo.NewRepositoryManager(cfg.Database)
	if err != nil {
		utils.Error("failed to initialize repository", "error", err)
		os.Exit(1)
	}
	defer repoManager.Close()

	// Initialize service and handler
	urlService := urlshortener.NewService(repoManager.URLRepo, cfg.Server.BaseURL)
	urlHandler := urlshortener.NewHandler(urlService)
	healthHandler := services.NewHealthHandler(repoManager)

	// Setup router
	router := urlshortener.SetupRouter(urlHandler, healthHandler)

	// Start server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		utils.Info("shutting down server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		server.Shutdown(shutdownCtx)
	}()

	utils.Info("server started", "port", cfg.Server.Port, "base_url", cfg.Server.BaseURL)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		utils.Error("server error", "error", err)
		os.Exit(1)
	}

	utils.Info("server stopped")
}
