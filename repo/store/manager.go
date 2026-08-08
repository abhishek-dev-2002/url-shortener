package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"github.com/abhishekmaurya/url-shortner/utils"
)

// Manager handles database connection lifecycle and provides access to stores.
type Manager struct {
	db *sql.DB
}

// NewManager creates a new store manager, establishes the DB connection,
// and runs migrations. This is the single place where DB logic lives.
func NewManager(cfg utils.DatabaseConfig) (*Manager, error) {
	db, err := sql.Open("postgres", cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	// Verify connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	utils.Info("connected to database")

	// Run migrations
	urlStore := NewURLShortenerStore(db)
	if err := urlStore.Migrate(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	utils.Info("database migrated")

	return &Manager{db: db}, nil
}

// DB returns the underlying sql.DB for health checks or other infra needs.
func (m *Manager) DB() *sql.DB {
	return m.db
}

func (m *Manager) URLStore() *URLShortenerStore {
	return NewURLShortenerStore(m.db)
}

// Close closes the database connection.
func (m *Manager) Close() error {
	return m.db.Close()
}
