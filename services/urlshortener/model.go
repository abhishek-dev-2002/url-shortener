package urlshortener

import "github.com/abhishekmaurya/url-shortner/models"

type shortenInput struct {
	OriginalURL string
	Alias       string
}

type shortenOutput struct {
	ShortCode   string `json:"short_code"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func getShortenRequest(body any) *models.ShortenRequest {
	return body.(*models.ShortenRequest)
}
