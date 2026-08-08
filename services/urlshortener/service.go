package urlshortener

import (
"context"
"net/http"
"strings"

"github.com/abhishekmaurya/url-shortner/models"
"github.com/abhishekmaurya/url-shortner/repo/repointerfaces"
"github.com/abhishekmaurya/url-shortner/services"
"github.com/abhishekmaurya/url-shortner/utils"
"github.com/abhishekmaurya/url-shortner/validator"
"github.com/gorilla/mux"
)

type Service struct {
	repo    repointerfaces.URLStore
	baseURL string
	codeGen *CodeGenerator
}

func NewService(repository repointerfaces.URLStore, baseURL string) *Service {
	return &Service{
		repo:    repository,
		baseURL: strings.TrimRight(baseURL, "/"),
		codeGen: NewCodeGenerator(repository),
	}
}

func (s *Service) Shorten(r *http.Request) (any, error) {
	return s.shorten(r.Context(), getShortenRequest(services.GetRequestBody(r)))
}

func (s *Service) Redirect(r *http.Request) (any, error) {
	code := mux.Vars(r)["code"]
	if code == "" {
		return nil, utils.BadRequest("short code is required")
	}

	originalURL, appErr := s.resolve(r.Context(), code)
	if appErr != nil {
		return nil, appErr
	}

	return &utils.RedirectResponse{
		Request:    r,
		URL:        originalURL,
		StatusCode: http.StatusMovedPermanently,
	}, nil
}

// shorten handles the core business logic for URL shortening.
// Duplicate URL policy: without alias → idempotent (returns existing code).
// With alias → always creates new mapping.
func (s *Service) shorten(ctx context.Context, req *models.ShortenRequest) (*shortenOutput, *utils.AppError) {
	originalURL, appErr := validator.ValidateURL(req.URL)
	if appErr != nil {
		return nil, appErr
	}

	input := toShortenInput(req, originalURL)

	if req.Alias != "" {
		return s.shortenWithAlias(ctx, input)
	}

	existing, err := s.repo.GetByOriginalURL(ctx, input.OriginalURL)
	if err != nil {
		utils.Error("failed to check existing url", "error", err)
		return nil, utils.InternalError("failed to process request")
	}
	if existing != nil {
		return toShortenOutput(s.baseURL, existing), nil
	}

	code, err := s.codeGen.Generate(ctx)
	if err != nil {
		utils.Error("failed to generate short code", "error", err)
		return nil, utils.InternalError("failed to generate short code")
	}

	created, err := s.repo.CreateURL(ctx, toURLModel(input, code, false))
	if err != nil {
		utils.Error("failed to create url", "error", err)
		return nil, utils.InternalError("failed to create short url")
	}

	utils.Info("url shortened", "short_code", created.ShortCode, "original_url", created.OriginalURL)
	return toShortenOutput(s.baseURL, created), nil
}

func (s *Service) resolve(ctx context.Context, code string) (string, *utils.AppError) {
	urlModel, err := s.repo.GetByShortCode(ctx, code)
	if err != nil {
		utils.Error("failed to resolve short code", "error", err, "code", code)
		return "", utils.InternalError("failed to resolve short code")
	}
	if urlModel == nil {
		return "", utils.NotFound("short code not found")
	}

	// Fire-and-forget: don't block redirect for analytics
go func() {
_ = s.repo.IncrementClickCount(context.Background(), code)
}()

return urlModel.OriginalURL, nil
}

func (s *Service) shortenWithAlias(ctx context.Context, input shortenInput) (*shortenOutput, *utils.AppError) {
if appErr := validator.ValidateAlias(input.Alias); appErr != nil {
return nil, appErr
}

exists, err := s.repo.ShortCodeExists(ctx, input.Alias)
if err != nil {
utils.Error("failed to check alias existence", "error", err, "alias", input.Alias)
return nil, utils.InternalError("failed to check alias availability")
}
if exists {
return nil, utils.Conflict("custom alias is already taken")
}

created, err := s.repo.CreateURL(ctx, toURLModel(input, input.Alias, true))
if err != nil {
utils.Error("failed to create url with alias", "error", err, "alias", input.Alias)
return nil, utils.InternalError("failed to create short url with alias")
}

utils.Info("url shortened with alias", "alias", input.Alias, "original_url", input.OriginalURL)
return toShortenOutput(s.baseURL, created), nil
}
