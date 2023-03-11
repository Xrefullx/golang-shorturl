package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/Xrefullx/golang-shorturl/internal/model"
	"github.com/Xrefullx/golang-shorturl/internal/storage"
	"github.com/google/uuid"
)

var _ URLShortener = (*ShortURLService)(nil)

type ShortURLService struct {
	db storage.Storage
}

func NewShortURLService(db storage.Storage) (*ShortURLService, error) {
	if db == nil {
		return nil, errors.New("ошибка инициализации хранилища")
	}

	return &ShortURLService{
		db: db,
	}, nil
}

func (sh *ShortURLService) DeleteURLList(userID uuid.UUID, shotIDList ...string) error {
	return sh.db.URL().DeleteURLBatch(userID, shotIDList...)
}

func (sh *ShortURLService) SaveURLList(src map[string]string, userID uuid.UUID) error {

	toAdd := make(map[string]model.ShortURL, len(src))

	checkShortID := make(map[string]string, len(src))

	for k, v := range src {

		sht := model.NewShortURL(v, userID)

		shortID, err := sh.genShortURL(v, sht.ID, checkShortID)
		if err != nil {
			return err
		}

		sht.ShortID = shortID
		if err := sh.db.URL().SaveURLBuff(&sht); err != nil {
			return err
		}
		toAdd[k] = sht
	}

	if err := sh.db.URL().SaveURLBuffFlush(); err != nil {
		return err
	}

	for k, v := range toAdd {
		src[k] = v.ShortID
	}

	return nil
}

func (sh *ShortURLService) GetUserURLList(ctx context.Context, userID uuid.UUID) ([]model.ShortURL, error) {
	list, err := sh.db.URL().GetUserURLList(ctx, userID, 100)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (sh *ShortURLService) GetURL(ctx context.Context, shortID string) (model.ShortURL, error) {
	longURL, err := sh.db.URL().GetURL(ctx, shortID)
	if err != nil {
		return model.ShortURL{}, err
	}

	return longURL, nil
}

func (sh *ShortURLService) SaveURL(ctx context.Context, srcURL string, userID uuid.UUID) (string, error) {

	sht := model.NewShortURL(srcURL, userID)

	var err error
	if sht.ShortID, err = sh.genShortURL(srcURL, sht.ID, nil); err != nil {
		return "", err
	}

	sht, err = sh.db.URL().SaveURL(ctx, sht)
	if err != nil {
		return "", err
	}

	return sht.ShortID, nil
}

func (sh *ShortURLService) Ping(ctx context.Context) error {
	return sh.db.Ping()
}

func (sh *ShortURLService) genShortURL(srcURL string, id uuid.UUID, generatedCheck map[string]string) (string, error) {
	shortID, err := sh.iterShortURLGenerator(string(srcURL), 0, id.String(), generatedCheck)
	if err != nil {
		return "", err
	}

	return shortID, nil
}

func (sh *ShortURLService) iterShortURLGenerator(srcURL string, iterationCount int, salt string, generatedCheck map[string]string) (string, error) {
	maxIterate := 10
	shortID := GenerateLink(srcURL, salt)

	existInCheck := false
	if generatedCheck != nil {
		_, existInCheck = generatedCheck[shortID]
	}
	exist, err := sh.db.URL().Exist(shortID)
	if err != nil {
		return "", fmt.Errorf("ошибка генерации короткой ссылки:%w", err)
	}
	if exist || existInCheck {
		iterationCount++
		if iterationCount > maxIterate {
			return "", fmt.Errorf("ошибка генерации короткой ссылки, число попыток:%v", maxIterate)
		}

		salt := uuid.New().String()

		shortID, err := sh.iterShortURLGenerator(srcURL, iterationCount, salt, generatedCheck)
		if err != nil {
			return "", err
		}

		return shortID, nil
	}

	if generatedCheck != nil {
		generatedCheck[shortID] = ""
	}

	return shortID, nil
}
