package main

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/abhishekmaurya/url-shortner/repo"
	"github.com/abhishekmaurya/url-shortner/services"
	"github.com/abhishekmaurya/url-shortner/services/urlshortener"
)

func setupRouting(ctx context.Context, repoMgr *repo.RepositoryManager, baseURL string) http.Handler {
	r := mux.NewRouter()
	r.Use(services.PanicRecoveryMiddleware)
	r.Use(services.CommonMiddleware)
	r.Use(services.RequestIDMiddleware)
	r.Use(services.RequestLoggerMiddleware)

	v1 := r.PathPrefix("/api/v1").Subrouter()
	urlshortener.SetupURLShortenerRouting(ctx, repoMgr, baseURL, v1, r)
	setupHealthRouting(r.PathPrefix("/health").Subrouter(), repoMgr)

	return r
}

func setupHealthRouting(r *mux.Router, repoMgr *repo.RepositoryManager) {
	healthService := services.NewHealthService(repoMgr)
	r.Methods(http.MethodGet).Path("").Handler(
		services.GenerateHandlerFunc(repoMgr, healthService.HealthCheck),
	)

	r.Methods(http.MethodGet).Path("/check").Handler(
		services.GenerateHandlerFunc(repoMgr, func(r *http.Request) (any, error) {
			return services.GetMessagePayload("ok"), nil
		}),
	)
}
