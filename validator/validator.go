package validator

import (
	"net/url"
	"strings"

	"github.com/abhishekmaurya/url-shortner/utils"
)

const base62Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// ValidateURL checks that a URL is well-formed and uses http or https.
// Returns the normalized URL string or an AppError.
func ValidateURL(rawURL string) (string, *utils.AppError) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", utils.BadRequest("url is required")
	}

	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return "", utils.BadRequest("invalid URL format")
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", utils.BadRequest("URL must use http or https scheme")
	}

	if parsed.Host == "" {
		return "", utils.BadRequest("URL must have a valid host")
	}

	return parsed.String(), nil
}

// ValidateAlias checks that a custom alias is between 3-20 characters and Base62 only.
func ValidateAlias(alias string) *utils.AppError {
	if len(alias) < utils.AliasMinLength || len(alias) > utils.AliasMaxLength {
		return utils.BadRequest("alias must be 3-20 characters long")
	}
	for _, c := range alias {
		if !strings.ContainsRune(base62Alphabet, c) {
			return utils.BadRequest("alias must contain only alphanumeric characters (a-z, A-Z, 0-9)")
		}
	}
	return nil
}
