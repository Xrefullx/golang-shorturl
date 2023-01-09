package file

import (
	"context"
	"errors"
	"github.com/Xrefullx/golang-shorturl/internal/storage/postgres/schema_postgres"

	"github.com/Xrefullx/golang-shorturl/internal/model"
	"github.com/Xrefullx/golang-shorturl/internal/storage"
	"github.com/google/uuid"
)

var _ storage.UserRepository = (*userRepository)(nil)

type userRepository struct {
	cache *cache
}

func newUserRepository(c *cache) (*userRepository, error) {
	if c == nil {
		return nil, errors.New("cant init repository cache not init")
	}

	return &userRepository{
		cache: c,
	}, nil
}

func (r *userRepository) Exist(userID uuid.UUID) (bool, error) {
	r.cache.RLock()
	_, ok := r.cache.userCache[userID]
	defer r.cache.RUnlock()

	return ok, nil
}

func (r *userRepository) AddUser(_ context.Context, user model.User) (model.User, error) {
	dbObj, err := schema_postgres.NewUserFromCanonical(user)
	if err != nil {
		return model.User{}, err
	}
	r.cache.Lock()
	r.cache.userCache[user.ID] = dbObj.ID
	defer r.cache.Unlock()

	return user, nil
}
