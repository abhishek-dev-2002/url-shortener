package urlshortener

import (
	"fmt"

	"github.com/abhishekmaurya/url-shortner/models"
)

func toShortenInput(req *models.ShortenRequest, originalURL string) shortenInput {
	return shortenInput{
		OriginalURL: originalURL,
		Alias:       req.Alias,
	}
}

func toURLModel(input shortenInput, shortCode string, customAlias bool) *models.URL {
	return &models.URL{
		ShortCode:   shortCode,
		OriginalURL: input.OriginalURL,
		CustomAlias: customAlias,
	}
}

func toShortenOutput(baseURL string, urlModel *models.URL) *shortenOutput {
	return &shortenOutput{
		ShortCode:   urlModel.ShortCode,
		ShortURL:    fmt.Sprintf("%s/%s", baseURL, urlModel.ShortCode),
		OriginalURL: urlModel.OriginalURL,
	}
}
