package file

import (
	"context"
	"errors"
	"fmt"
	"github.com/Xrefullx/golang-shorturl/internal/storage/postgres/schema_postgres"

	"github.com/Xrefullx/golang-shorturl/internal/model"
	"github.com/Xrefullx/golang-shorturl/internal/shrterr"
	st "github.com/Xrefullx/golang-shorturl/internal/storage"
	"github.com/google/uuid"
)

var _ st.URLRepository = (*shortURLRepository)(nil)

type shortURLRepository struct {
	cache    *cache
	fileName string
}

func newShortURLRepository(c *cache, fileName string) (*shortURLRepository, error) {
	if c == nil {
		return nil, errors.New("cant init repository cache not init")
	}

	return &shortURLRepository{
		cache:    c,
		fileName: fileName,
	}, nil
}

func (r *shortURLRepository) DeleteURLBatch(userID uuid.UUID, shortIDList ...string) error {
	if len(shortIDList) == 0 {
		return nil
	}
	for _, v := range shortIDList {
		sht, _ := r.GetURL(context.TODO(), v)
		if sht != (model.ShortURL{}) {
			if sht.UserID == userID && !sht.IsDeleted {
				sht.IsDeleted = true
				toAdd, err := schema_postgres.NewURLFromCanonical(sht)
				if err != nil {
					return fmt.Errorf("ошибка обновления запси: %w", err)
				}

				r.cache.Lock()
				r.cache.urlCache[sht.ID] = toAdd
				r.cache.Unlock()
			}

		}
	}
	return nil
}

func (r *shortURLRepository) SaveURLBuff(sht *model.ShortURL) error {
	if sht == nil {
		return errors.New("short URL is nil")
	}

	dbObj, err := r.SaveURL(context.TODO(), *sht)
	if err != nil {
		return err
	}
	sht = &dbObj
	return nil
}

func (r *shortURLRepository) SaveURLBuffFlush() error {
	return nil
}

func (r *shortURLRepository) SaveURL(_ context.Context, sht model.ShortURL) (model.ShortURL, error) {

	dbObj, err := schema_postgres.NewURLFromCanonical(sht)
	if err != nil {
		return model.ShortURL{}, fmt.Errorf("ошибка хранилица:%w", err)
	}

	exist, _ := r.Exist(dbObj.ShortID)
	if exist {
		return model.ShortURL{}, errors.New("shortID уже существует")
	}

	existSrcURL, _ := r.ExistSrcURL(dbObj.URL)
	if existSrcURL {

		return model.ShortURL{}, &shrterr.ErrorConflictSaveURL{
			Err:           errors.New("конфликт добавления записи, URL уже существует"),
			ExistShortURL: r.GetShortURLBySrcURL(dbObj.URL),
		}
	}

	if userExist := r.UserExist(sht.UserID); !userExist {
		return model.ShortURL{}, errors.New("пользователь не найден")
	}

	r.cache.Lock()
	defer r.cache.Unlock()
	if r.fileName != "" {

		if err := r.writeToFile(dbObj); err != nil {
			return model.ShortURL{}, err
		}
	}

	r.cache.urlCache[dbObj.ID] = dbObj
	r.cache.shortURLidx[dbObj.ShortID] = dbObj.ID
	r.cache.srcURLidx[dbObj.URL] = dbObj.ID

	return sht, nil
}

func (r *shortURLRepository) GetURL(_ context.Context, shortID string) (model.ShortURL, error) {
	if shortID == "" {
		return model.ShortURL{}, errors.New("нельзя использовать пустой id")
	}

	r.cache.RLock()
	idx, ok := r.cache.shortURLidx[shortID]
	defer r.cache.RUnlock()

	if ok {
		item, ok := r.cache.urlCache[idx]
		if ok {
			return item.ToCanonical()
		}
	}

	return model.ShortURL{}, nil
}

func (r *shortURLRepository) GetShortURLBySrcURL(url string) string {
	r.cache.RLock()
	id, ok := r.cache.srcURLidx[url]
	defer r.cache.RUnlock()

	if ok {
		sht, okSht := r.cache.urlCache[id]
		if okSht {
			return sht.ShortID
		}
	}

	return ""
}

func (r *shortURLRepository) GetUserURLList(_ context.Context, userID uuid.UUID, limit int) ([]model.ShortURL, error) {
	r.cache.RLock()
	defer r.cache.RUnlock()

	if len(r.cache.urlCache) == 0 {
		return nil, nil
	}
	var dbURLs schema_postgres.URLList
	dbURLs = make([]schema_postgres.ShortURL, 0, limit)
	for _, v := range r.cache.urlCache {
		if v.UserID != uuid.Nil && v.UserID == userID {
			dbURLs = append(dbURLs, v)
			if len(dbURLs) == limit {
				break
			}
		}
	}

	if len(dbURLs) == 0 {
		return nil, nil
	}

	return dbURLs.ToCanonical()
}

func (r *shortURLRepository) UserExist(userID uuid.UUID) bool {
	r.cache.RLock()
	_, userExist := r.cache.userCache[userID]
	defer r.cache.RUnlock()
	return userExist
}

func (r *shortURLRepository) Exist(shortID string) (bool, error) {
	r.cache.RLock()
	_, ok := r.cache.shortURLidx[shortID]
	defer r.cache.RUnlock()

	return ok, nil
}

func (r *shortURLRepository) ExistSrcURL(url string) (bool, error) {
	r.cache.RLock()
	_, ok := r.cache.srcURLidx[url]
	defer r.cache.RUnlock()

	return ok, nil
}

func (r *shortURLRepository) writeToFile(sht schema_postgres.ShortURL) error {
	fileWriter, err := newFileWriter(r.fileName)
	if err != nil {
		return fmt.Errorf("ошибка записи в хранилище: %w", err)
	}

	defer fileWriter.Close()
	if err := fileWriter.WriteURL(sht); err != nil {
		return fmt.Errorf("ошибка записи в хранилище: %w", err)
	}

	return nil
}
