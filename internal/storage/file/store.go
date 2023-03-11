package file

import (
	"errors"
	"fmt"

	"github.com/Xrefullx/golang-shorturl/internal/storage"
	"github.com/google/uuid"
)

var _ storage.Storage = (*Storage)(nil)

type Storage struct {
	shortURLRepo *shortURLRepository
	userRepo     *userRepository
	fileName     string
	cache        *cache
}

func NewFileStorage(fileName string) (*Storage, error) {
	st := Storage{
		fileName: fileName,
		cache:    newCache(),
	}

	var err error

	st.shortURLRepo, err = newShortURLRepository(st.cache, st.fileName)
	if err != nil {
		return nil, fmt.Errorf("ошибка инициализации хранилища: %w", err)
	}

	st.userRepo, err = newUserRepository(st.cache)
	if err != nil {
		return nil, fmt.Errorf("ошибка инициализации хранилища: %w", err)
	}

	if st.fileName != "" {
		if err := st.initFromFile(); err != nil {
			return nil, fmt.Errorf("ошибка инициализации хранилища: %w", err)
		}
	}

	return &st, nil
}

func (s *Storage) URL() storage.URLRepository {
	if s.shortURLRepo != nil {
		return s.shortURLRepo
	}
	return s.shortURLRepo
}

func (s *Storage) User() storage.UserRepository {
	if s.shortURLRepo != nil {
		return s.userRepo
	}
	return s.userRepo
}

func (s *Storage) initFromFile() error {
	fileReader, err := newFileReader(s.fileName)
	if err != nil {
		return fmt.Errorf("ошибка чтения из хранилища: %w", err)
	}

	data, err := fileReader.ReadAll()
	defer fileReader.Close()
	if err != nil {
		return fmt.Errorf("ошибка чтения из хранилища: %w", err)
	}

	s.cache.urlCache = data

	if len(data) > 0 {
		for _, v := range data {
			existShortID, _ := s.shortURLRepo.Exist(v.ShortID)
			if !existShortID {
				s.cache.shortURLidx[v.ShortID] = v.ID
			}

			existUser, _ := s.userRepo.Exist(v.UserID)
			if v.UserID != uuid.Nil && !existUser {
				s.cache.userCache[v.UserID] = v.UserID
			}

			existURL, _ := s.shortURLRepo.Exist(v.URL)
			if !existURL {
				s.cache.srcURLidx[v.URL] = v.ID
			}
		}
	}

	return nil
}

func (s *Storage) Ping() error {
	return errors.New("db not initialized")
}

func (s *Storage) Close() {}
