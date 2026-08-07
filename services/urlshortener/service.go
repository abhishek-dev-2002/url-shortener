package urlshortener

import (
"context"
"fmt"
"strings"

"github.com/abhishekmaurya/url-shortner/models"
"github.com/abhishekmaurya/url-shortner/repo/repointerfaces"
"github.com/abhishekmaurya/url-shortner/utils"
"github.com/abhishekmaurya/url-shortner/validator"
)

// Service handles business logic for URL shortening.
type Service struct {
	repo    repointerfaces.URLRepository
	baseURL string
	codeGen *CodeGenerator
}

// NewService creates a new URL shortener service.
func NewService(repository repointerfaces.URLRepository, baseURL string) *Service {
	return &Service{
		repo:    repository,
		baseURL: strings.TrimRight(baseURL, "/"),
		codeGen: NewCodeGenerator(repository),
	}
}

// Shorten creates a short URL for the given original URL.
//
// Duplicate URL policy (deliberate design decision):
//   - Without alias: idempotent — returns existing short code for the same URL.
//   - With alias: always creates a new mapping (one URL can have multiple aliases).
func (s *Service) Shorten(ctx context.Context, req models.ShortenRequest) (*models.ShortenResponse, *utils.AppError) {
	originalURL, appErr := validator.ValidateURL(req.URL)
	if appErr != nil {
		return nil, appErr
	}

	// Custom alias flow
	if req.Alias != "" {
		return s.shortenWithAlias(ctx, originalURL, req.Alias)
	}

	// Idempotent: return existing short code if URL was already shortened
	existing, err := s.repo.GetByOriginalURL(ctx, originalURL)
	if err != nil {
		utils.Error("failed to check existing url", "error", err)
		return nil, utils.InternalError("failed to process request")
	}
	if existing != nil {
		return s.buildResponse(existing), nil
	}

	// Generate unique code from pre-allocated block (no collision possible)
	code, err := s.codeGen.Generate(ctx)
	if err != nil {
		utils.Error("failed to generate short code", "error", err)
		return nil, utils.InternalError("failed to generate short code")
	}

	urlModel := &models.URL{
		ShortCode:   code,
		OriginalURL: originalURL,
		CustomAlias: false,
	}

	created, err := s.repo.CreateURL(ctx, urlModel)
	if err != nil {
		utils.Error("failed to create url", "error", err)
		return nil, utils.InternalError("failed to create short url")
	}

	utils.Info("url shortened", "short_code", created.ShortCode, "original_url", created.OriginalURL)
	return s.buildResponse(created), nil
}

// Resolve looks up a short code and returns the original URL.
func (s *Service) Resolve(ctx context.Context, code string) (string, *utils.AppError) {
	urlModel, err := s.repo.GetByShortCode(ctx, code)
	if err != nil {
		utils.Error("failed to resolve short code", "error", err, "code", code)
		return "", utils.InternalError("failed to resolve short code")
	}
	if urlModel == nil {
		return "", utils.NotFound("short code not found")
	}

	// Increment click count (fire-and-forget)
	go func() {
		_ = s.repo.IncrementClickCount(context.Background(), code)
	}()

	return urlModel.OriginalURL, nil
}

func (s *Service) shortenWithAlias(ctx context.Context, originalURL, alias string) (*models.ShortenResponse, *utils.AppError) {
	if appErr := validator.ValidateAlias(alias); appErr != nil {
		return nil, appErr
	}

	exists, err := s.repo.ShortCodeExists(ctx, alias)
	if err != nil {
		utils.Error("failed to check alias existence", "error", err, "alias", alias)
		return nil, utils.InternalError("failed to check alias availability")
	}
	if exists {
		return nil, utils.Conflict("custom alias is already taken")
	}

	urlModel := &models.URL{
		ShortCode:   alias,
		OriginalURL: originalURL,
		CustomAlias: true,
	}

	created, err := s.repo.CreateURL(ctx, urlModel)
	if err != nil {
		utils.Error("failed to create url with alias", "error", err, "alias", alias)
		return nil, utils.InternalError("failed to create short url with alias")
	}

	utils.Info("url shortened with alias", "alias", alias, "original_url", originalURL)
	return s.buildResponse(created), nil
}

func (s *Service) buildResponse(urlModel *models.URL) *models.ShortenResponse {
	return &models.ShortenResponse{
		ShortCode:   urlModel.ShortCode,
		ShortURL:    fmt.Sprintf("%s/%s", s.baseURL, urlModel.ShortCode),
		OriginalURL: urlModel.OriginalURL,
	}
}
