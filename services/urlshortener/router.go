package urlshortener

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/abhishekmaurya/url-shortner/models"
	"github.com/abhishekmaurya/url-shortner/repo"
	"github.com/abhishekmaurya/url-shortner/services"
)

// SetupURLShortenerRouting registers URL shortener endpoints.
// Shorten lives under /api/v1, redirect lives at root level for clean short URLs.
func SetupURLShortenerRouting(_ context.Context, repoMgr *repo.RepositoryManager, baseURL string, v1Router *mux.Router, rootRouter *mux.Router) {
	service := NewService(repoMgr.GetURLStore(), baseURL)

	// POST /api/v1/shorten
	v1Router.Methods(http.MethodPost).Path("/shorten").Handler(
		services.DecoderMiddleware(func() any {
			return &models.ShortenRequest{}
		}, services.GenerateHandlerFunc(repoMgr, service.Shorten)),
	)

	// GET /{code} — at root level for clean redirect URLs
	rootRouter.Methods(http.MethodGet).Path("/{code}").Handler(
		services.GenerateHandlerFunc(repoMgr, service.Redirect),
	)
}
