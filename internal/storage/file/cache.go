package file

import (
	"github.com/Xrefullx/golang-shorturl/internal/storage/postgres/schema_postgres"
	"sync"

	"github.com/google/uuid"
)

type cache struct {
	sync.RWMutex
	urlCache    map[uuid.UUID]schema_postgres.ShortURL
	shortURLidx map[string]uuid.UUID
	srcURLidx   map[string]uuid.UUID
	userCache   map[uuid.UUID]uuid.UUID
}

func newCache() *cache {
	return &cache{
		urlCache:    make(map[uuid.UUID]schema_postgres.ShortURL),
		userCache:   make(map[uuid.UUID]uuid.UUID),
		shortURLidx: make(map[string]uuid.UUID),
		srcURLidx:   make(map[string]uuid.UUID),
	}
}
