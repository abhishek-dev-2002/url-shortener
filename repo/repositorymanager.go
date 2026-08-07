package repo

import (
"database/sql"

"github.com/abhishekmaurya/url-shortner/repo/repointerfaces"
"github.com/abhishekmaurya/url-shortner/repo/store"
"github.com/abhishekmaurya/url-shortner/utils"
)

// RepositoryManager holds all repository instances and the store manager.
// This is the abstraction layer that services depend on.
type RepositoryManager struct {
	storeManager *store.Manager
	URLRepo      repointerfaces.URLRepository
}

// NewRepositoryManager initializes the store manager (DB connection, migrations)
// and wires up all repository implementations.
func NewRepositoryManager(cfg utils.DatabaseConfig) (*RepositoryManager, error) {
	storeMgr, err := store.NewManager(cfg)
	if err != nil {
		return nil, err
	}

	return &RepositoryManager{
		storeManager: storeMgr,
		URLRepo:      storeMgr.URLStore(),
	}, nil
}

// DB exposes the underlying DB connection for health checks.
func (rm *RepositoryManager) DB() *sql.DB {
	return rm.storeManager.DB()
}

// Close closes all store connections.
func (rm *RepositoryManager) Close() error {
	return rm.storeManager.Close()
}
