package urlshortener

import (
	"context"
	"sync"
	"testing"

	"github.com/abhishekmaurya/url-shortner/models"
	
)

// mockRepo is a minimal in-memory repository for testing.
type mockRepo struct {
	mu      sync.RWMutex
	urls    map[string]*models.URL
	nextID  int64
	blockID int64
}

func newMockRepo() *mockRepo {
	return &mockRepo{urls: make(map[string]*models.URL), blockID: 1}
}

func (m *mockRepo) CreateURL(_ context.Context, url *models.URL) (*models.URL, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextID++
	url.ID = m.nextID
	stored := *url
	m.urls[url.ShortCode] = &stored
	return &stored, nil
}

func (m *mockRepo) GetByShortCode(_ context.Context, code string) (*models.URL, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	url, ok := m.urls[code]
	if !ok {
		return nil, nil
	}
	result := *url
	return &result, nil
}

func (m *mockRepo) GetByOriginalURL(_ context.Context, originalURL string) (*models.URL, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, url := range m.urls {
		if url.OriginalURL == originalURL && !url.CustomAlias {
			result := *url
			return &result, nil
		}
	}
	return nil, nil
}

func (m *mockRepo) ShortCodeExists(_ context.Context, code string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.urls[code]
	return exists, nil
}

func (m *mockRepo) IncrementClickCount(_ context.Context, code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if url, ok := m.urls[code]; ok {
		url.ClickCount++
	}
	return nil
}

func (m *mockRepo) AllocateIDBlock(_ context.Context, blockSize int64) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	start := m.blockID
	m.blockID += blockSize
	return start, nil
}

// --- Tests ---

func TestShorten_ValidURL(t *testing.T) {
	mockRepository := newMockRepo()
	svc := NewService(mockRepository, "http://localhost:8080")

	resp, appErr := svc.shorten(context.Background(), &models.ShortenRequest{URL: "https://example.com/path"})
	if appErr != nil {
		t.Fatalf("unexpected error: %s", appErr.Message)
	}
	if resp.ShortCode == "" {
		t.Error("expected non-empty short code")
	}
	if resp.OriginalURL != "https://example.com/path" {
		t.Errorf("expected original url preserved, got %s", resp.OriginalURL)
	}
	if resp.ShortURL != "http://localhost:8080/"+resp.ShortCode {
		t.Errorf("unexpected short url: %s", resp.ShortURL)
	}
}

func TestShorten_DuplicateURL_ReturnsExisting(t *testing.T) {
	mockRepository := newMockRepo()
	svc := NewService(mockRepository, "http://localhost:8080")
	ctx := context.Background()

	resp1, _ := svc.shorten(ctx, &models.ShortenRequest{URL: "https://example.com/dup"})
	resp2, _ := svc.shorten(ctx, &models.ShortenRequest{URL: "https://example.com/dup"})

	if resp1.ShortCode != resp2.ShortCode {
		t.Errorf("expected same code for duplicate URL, got %s and %s", resp1.ShortCode, resp2.ShortCode)
	}
}

func TestShorten_CustomAlias(t *testing.T) {
	mockRepository := newMockRepo()
	svc := NewService(mockRepository, "http://localhost:8080")

	resp, appErr := svc.shorten(context.Background(), &models.ShortenRequest{
		URL:   "https://example.com",
		Alias: "mysite",
	})
	if appErr != nil {
		t.Fatalf("unexpected error: %s", appErr.Message)
	}
	if resp.ShortCode != "mysite" {
		t.Errorf("expected 'mysite', got %s", resp.ShortCode)
	}
}

func TestShorten_DuplicateAlias_ReturnsConflict(t *testing.T) {
	mockRepository := newMockRepo()
	svc := NewService(mockRepository, "http://localhost:8080")
	ctx := context.Background()

	svc.shorten(ctx, &models.ShortenRequest{URL: "https://a.com", Alias: "taken"})
	_, appErr := svc.shorten(ctx, &models.ShortenRequest{URL: "https://b.com", Alias: "taken"})

	if appErr == nil || appErr.Code != "CONFLICT" {
		t.Errorf("expected CONFLICT error, got %v", appErr)
	}
}

func TestShorten_InvalidURL(t *testing.T) {
	mockRepository := newMockRepo()
	svc := NewService(mockRepository, "http://localhost:8080")

	cases := []string{"", "not-a-url", "ftp://x.com", "javascript:alert(1)"}
	for _, url := range cases {
		_, appErr := svc.shorten(context.Background(), &models.ShortenRequest{URL: url})
		if appErr == nil || appErr.Code != "BAD_REQUEST" {
			t.Errorf("URL %q: expected BAD_REQUEST, got %v", url, appErr)
		}
	}
}

func TestShorten_InvalidAlias(t *testing.T) {
	mockRepository := newMockRepo()
	svc := NewService(mockRepository, "http://localhost:8080")

	cases := []string{"ab", "a", "abc-def", "has space", "toolongaliasthatisfar!!"}
	for _, alias := range cases {
		_, appErr := svc.shorten(context.Background(), &models.ShortenRequest{
			URL: "https://example.com", Alias: alias,
		})
		if appErr == nil || appErr.Code != "BAD_REQUEST" {
			t.Errorf("alias %q: expected BAD_REQUEST, got %v", alias, appErr)
		}
	}
}

func TestResolve_ExistingCode(t *testing.T) {
	mockRepository := newMockRepo()
	svc := NewService(mockRepository, "http://localhost:8080")
	ctx := context.Background()

	resp, _ := svc.shorten(ctx, &models.ShortenRequest{URL: "https://example.com/resolve"})
	originalURL, appErr := svc.resolve(ctx, resp.ShortCode)

	if appErr != nil {
		t.Fatalf("unexpected error: %s", appErr.Message)
	}
	if originalURL != "https://example.com/resolve" {
		t.Errorf("expected https://example.com/resolve, got %s", originalURL)
	}
}

func TestResolve_NotFound(t *testing.T) {
	mockRepository := newMockRepo()
	svc := NewService(mockRepository, "http://localhost:8080")

	_, appErr := svc.resolve(context.Background(), "nonexist")
	if appErr == nil || appErr.Code != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND, got %v", appErr)
	}
}

func TestShorten_GeneratesUniqueCodes(t *testing.T) {
	mockRepository := newMockRepo()
	svc := NewService(mockRepository, "http://localhost:8080")
	ctx := context.Background()
	seen := make(map[string]bool)

	for i := 0; i < 100; i++ {
		resp, appErr := svc.shorten(ctx, &models.ShortenRequest{
			URL: "https://example.com/" + string(rune('a'+i%26)) + string(rune('0'+i/26)),
		})
		if appErr != nil {
			t.Fatalf("unexpected error on iter %d: %s", i, appErr.Message)
		}
		if seen[resp.ShortCode] {
			t.Fatalf("duplicate code generated: %s", resp.ShortCode)
		}
		seen[resp.ShortCode] = true
	}
}

func TestShorten_AliasAllowsSameURLWithDifferentCodes(t *testing.T) {
	mockRepository := newMockRepo()
	svc := NewService(mockRepository, "http://localhost:8080")
	ctx := context.Background()

	resp1, _ := svc.shorten(ctx, &models.ShortenRequest{URL: "https://example.com/multi"})
	resp2, _ := svc.shorten(ctx, &models.ShortenRequest{URL: "https://example.com/multi", Alias: "custom"})

	if resp1.ShortCode == resp2.ShortCode {
		t.Error("expected different codes for aliased vs auto-generated")
	}
	if resp2.ShortCode != "custom" {
		t.Errorf("expected 'custom', got %s", resp2.ShortCode)
	}
}
