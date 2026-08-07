package urlshortener

import (
"context"
"net/http"

"github.com/gin-gonic/gin"

"github.com/abhishekmaurya/url-shortner/models"
"github.com/abhishekmaurya/url-shortner/utils"
)

// URLService is the interface the handler depends on.
// Defined here (consumer side) for loose coupling.
type URLService interface {
	Shorten(ctx context.Context, req models.ShortenRequest) (*models.ShortenResponse, *utils.AppError)
	Resolve(ctx context.Context, code string) (string, *utils.AppError)
}

// Handler handles HTTP requests for URL shortening.
type Handler struct {
	service URLService
}

// NewHandler creates a new URL handler.
func NewHandler(service URLService) *Handler {
	return &Handler{service: service}
}

// Shorten handles POST /api/v1/shorten.
func (h *Handler) Shorten(c *gin.Context) {
	var req models.ShortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, utils.BadRequest("invalid request body"))
		return
	}

	resp, appErr := h.service.Shorten(c.Request.Context(), req)
	if appErr != nil {
		utils.SendError(c, appErr)
		return
	}

	utils.SendSuccess(c, http.StatusCreated, resp, "SUCCESS")
}

// Redirect handles GET /:code.
func (h *Handler) Redirect(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		utils.SendError(c, utils.BadRequest("short code is required"))
		return
	}

	originalURL, appErr := h.service.Resolve(c.Request.Context(), code)
	if appErr != nil {
		utils.SendError(c, appErr)
		return
	}

	c.Redirect(http.StatusMovedPermanently, originalURL)
}
