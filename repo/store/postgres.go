package store

import (
"context"
"database/sql"
"fmt"
"time"

"github.com/abhishekmaurya/url-shortner/models"
"github.com/abhishekmaurya/url-shortner/utils"
)

// PostgresURLStore implements repointerfaces.URLRepository using PostgreSQL.
type PostgresURLStore struct {
	db *sql.DB
}

// NewPostgresURLStore creates a new PostgreSQL URL store.
func NewPostgresURLStore(db *sql.DB) *PostgresURLStore {
	return &PostgresURLStore{db: db}
}

// Migrate runs the database schema migration.
func (s *PostgresURLStore) Migrate(ctx context.Context) error {
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

		-- Sequence for short code block allocation.
		-- Each call to nextval advances by 1; we call setval to jump by blockSize.
		CREATE SEQUENCE IF NOT EXISTS shortcode_id_seq START WITH 1 INCREMENT BY 1;
	`
	_, err := s.db.ExecContext(ctx, query)
	if err != nil {
		utils.Error("migration failed", "error", err)
	}
	return err
}

// AllocateIDBlock atomically allocates a block of sequential IDs from the DB sequence.
// Returns the start of the block. The caller owns IDs [start, start+blockSize).
//
// This is the key to the block allocation strategy:
//   - Only 1 DB call per 10,000 URLs (amortized cost ≈ 0)
//   - Multiple instances can allocate different blocks concurrently — no collisions
//   - The sequence is the single source of truth — atomic, crash-safe
func (s *PostgresURLStore) AllocateIDBlock(ctx context.Context, blockSize int64) (int64, error) {
	// setval(seq, current + blockSize) returns the new value.
	// We use a single atomic operation to claim [current+1, current+blockSize].
	query := `SELECT setval('shortcode_id_seq', nextval('shortcode_id_seq') + $1 - 1)`

	var endID int64
	err := s.db.QueryRowContext(ctx, query, blockSize).Scan(&endID)
	if err != nil {
		utils.Error("failed to allocate ID block", "error", err)
		return 0, fmt.Errorf("failed to allocate ID block: %w", err)
	}

	// endID is the last ID in the block. Start = endID - blockSize + 1
	startID := endID - blockSize + 1
	return startID, nil
}

func (s *PostgresURLStore) CreateURL(ctx context.Context, url *models.URL) (*models.URL, error) {
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

func (s *PostgresURLStore) GetByShortCode(ctx context.Context, shortCode string) (*models.URL, error) {
	query := `
		SELECT id, short_code, original_url, custom_alias, created_at, click_count
		FROM urls WHERE short_code = $1
	`

	var url models.URL
	err := s.db.QueryRowContext(ctx, query, shortCode).Scan(
&url.ID, &url.ShortCode, &url.OriginalURL,
		&url.CustomAlias, &url.CreatedAt, &url.ClickCount,
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

func (s *PostgresURLStore) GetByOriginalURL(ctx context.Context, originalURL string) (*models.URL, error) {
	query := `
		SELECT id, short_code, original_url, custom_alias, created_at, click_count
		FROM urls WHERE original_url = $1 AND custom_alias = FALSE
		ORDER BY created_at ASC LIMIT 1
	`

	var url models.URL
	err := s.db.QueryRowContext(ctx, query, originalURL).Scan(
&url.ID, &url.ShortCode, &url.OriginalURL,
		&url.CustomAlias, &url.CreatedAt, &url.ClickCount,
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

func (s *PostgresURLStore) ShortCodeExists(ctx context.Context, shortCode string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM urls WHERE short_code = $1)`

	var exists bool
	err := s.db.QueryRowContext(ctx, query, shortCode).Scan(&exists)
	if err != nil {
		utils.Error("failed to check short code existence", "error", err, "short_code", shortCode)
		return false, fmt.Errorf("failed to check short code existence: %w", err)
	}

	return exists, nil
}

func (s *PostgresURLStore) IncrementClickCount(ctx context.Context, shortCode string) error {
	query := `UPDATE urls SET click_count = click_count + 1 WHERE short_code = $1`

	_, err := s.db.ExecContext(ctx, query, shortCode)
	if err != nil {
		utils.Error("failed to increment click count", "error", err, "short_code", shortCode)
		return fmt.Errorf("failed to increment click count: %w", err)
	}

	return nil
}
