package app

import (
	"errors"
	"github.com/Xrefullx/golang-shorturl/internal/storage"
	"github.com/google/uuid"
)

type shortservice struct {
	db storage.URLStore
}

func NewShort(db storage.URLStore) (*shortservice, error) {
	return &shortservice{
		db: db,
	}, nil
}

func (ss *shortservice) Get(short string) (string, error) {
	long, err := ss.db.Get(short)
	if err != nil {
		return "", err
	}

	return long, nil
}
func (ss *shortservice) Save(search string) (string, error) {
	short, err := ss.generateShort(string(search))
	if err != nil {
		return "", err
	}

	_, err = ss.db.Save(short, string(search))
	if err != nil {
		return "", err
	}

	return short, nil
}

func (ss *shortservice) generateShort(search string) (string, error) {
	shortID, err := ss.iterShortGenerator(string(search), 0, "")
	if err != nil {
		return "", err
	}

	return shortID, nil
}

func (ss *shortservice) iterShortGenerator(search string, count int, encod string) (string, error) {
	short := GenerateLink(search, encod)
	if !ss.db.IsShort(short) {
		count++
		encod := uuid.New().String()

		short, err := ss.iterShortGenerator(search, count, encod)
		if err != nil || count > 10 {
			return "", errors.New("err generate")
		}

		return short, nil
	}

	return short, nil
}
