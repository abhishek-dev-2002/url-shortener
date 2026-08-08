package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/abhishekmaurya/url-shortner/models"
	"github.com/abhishekmaurya/url-shortner/utils"
)

type URLShortenerStore struct {
	db *sql.DB
}

func NewURLShortenerStore(db *sql.DB) *URLShortenerStore {
	return &URLShortenerStore{db: db}
}

func (s *URLShortenerStore) Migrate(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS urls (
			id BIGSERIAL PRIMARY KEY,
			short_code VARCHAR(20) UNIQUE NOT NULL,
			original_url TEXT NOT NULL,
			custom_alias BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			click_count BIGINT DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS idx_urls_short_code ON urls(short_code);
		CREATE INDEX IF NOT EXISTS idx_urls_original_url ON urls(original_url);
		CREATE SEQUENCE IF NOT EXISTS shortcode_id_seq START WITH 1 INCREMENT BY 1;
	`

	_, err := s.db.ExecContext(ctx, query)
	if err != nil {
		utils.Error("migration failed", "error", err)
	}
	return err
}

func (s *URLShortenerStore) AllocateIDBlock(ctx context.Context, blockSize int64) (int64, error) {
	query := `SELECT setval('shortcode_id_seq', nextval('shortcode_id_seq') + $1 - 1)`

	var endID int64
	if err := s.db.QueryRowContext(ctx, query, blockSize).Scan(&endID); err != nil {
		utils.Error("failed to allocate ID block", "error", err)
		return 0, fmt.Errorf("failed to allocate ID block: %w", err)
	}

	return endID - blockSize + 1, nil
}

func (s *URLShortenerStore) CreateURL(ctx context.Context, url *models.URL) (*models.URL, error) {
	query := `
		INSERT INTO urls (short_code, original_url, custom_alias, created_at, click_count)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	now := time.Now().UTC()
	err := s.db.QueryRowContext(ctx, query,
		url.ShortCode,
		url.OriginalURL,
		url.CustomAlias,
		now,
		0,
	).Scan(&url.ID, &url.CreatedAt)
	if err != nil {
		utils.Error("failed to create url", "error", err, "short_code", url.ShortCode)
		return nil, fmt.Errorf("failed to create url: %w", err)
	}

	return url, nil
}

func (s *URLShortenerStore) GetByShortCode(ctx context.Context, shortCode string) (*models.URL, error) {
	query := `
		SELECT id, short_code, original_url, custom_alias, created_at, click_count
		FROM urls WHERE short_code = $1
	`

	var url models.URL
	err := s.db.QueryRowContext(ctx, query, shortCode).Scan(
		&url.ID,
		&url.ShortCode,
		&url.OriginalURL,
		&url.CustomAlias,
		&url.CreatedAt,
		&url.ClickCount,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		utils.Error("failed to get url by short code", "error", err, "short_code", shortCode)
		return nil, fmt.Errorf("failed to get url by short code: %w", err)
	}

	return &url, nil
}

func (s *URLShortenerStore) GetByOriginalURL(ctx context.Context, originalURL string) (*models.URL, error) {
	query := `
		SELECT id, short_code, original_url, custom_alias, created_at, click_count
		FROM urls
		WHERE original_url = $1 AND custom_alias = FALSE
		ORDER BY created_at ASC
		LIMIT 1
	`

	var url models.URL
	err := s.db.QueryRowContext(ctx, query, originalURL).Scan(
		&url.ID,
		&url.ShortCode,
		&url.OriginalURL,
		&url.CustomAlias,
		&url.CreatedAt,
		&url.ClickCount,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		utils.Error("failed to get url by original url", "error", err)
		return nil, fmt.Errorf("failed to get url by original url: %w", err)
	}

	return &url, nil
}

func (s *URLShortenerStore) ShortCodeExists(ctx context.Context, shortCode string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM urls WHERE short_code = $1)`

	var exists bool
	if err := s.db.QueryRowContext(ctx, query, shortCode).Scan(&exists); err != nil {
		utils.Error("failed to check short code existence", "error", err, "short_code", shortCode)
		return false, fmt.Errorf("failed to check short code existence: %w", err)
	}

	return exists, nil
}

func (s *URLShortenerStore) IncrementClickCount(ctx context.Context, shortCode string) error {
	query := `UPDATE urls SET click_count = click_count + 1 WHERE short_code = $1`

	if _, err := s.db.ExecContext(ctx, query, shortCode); err != nil {
		utils.Error("failed to increment click count", "error", err, "short_code", shortCode)
		return fmt.Errorf("failed to increment click count: %w", err)
	}

	return nil
}
