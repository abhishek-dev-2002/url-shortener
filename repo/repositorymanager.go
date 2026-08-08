package repo

import (
"database/sql"

"github.com/abhishekmaurya/url-shortner/repo/repointerfaces"
"github.com/abhishekmaurya/url-shortner/repo/store"
"github.com/abhishekmaurya/url-shortner/utils"
)

// RepositoryManager owns the database lifecycle and exposes repository interfaces.
// Services depend on this — never on *sql.DB directly.
type RepositoryManager struct {
	storeManager *store.Manager
}

func NewRepositoryManager(cfg utils.DatabaseConfig) (*RepositoryManager, error) {
	storeMgr, err := store.NewManager(cfg)
	if err != nil {
		return nil, err
	}

	return &RepositoryManager{storeManager: storeMgr}, nil
}

// GetURLStore returns the URL repository (interface-typed for loose coupling).
func (rm *RepositoryManager) GetURLStore() repointerfaces.URLStore {
	return rm.storeManager.URLStore()
}

func (rm *RepositoryManager) DB() *sql.DB {
	return rm.storeManager.DB()
}

func (rm *RepositoryManager) Close() error {
	return rm.storeManager.Close()
}
