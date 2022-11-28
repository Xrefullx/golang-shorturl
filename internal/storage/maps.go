package storage

import (
	"errors"
	"github.com/Xrefullx/golang-shorturl/internal/app"
	"strconv"
)

var _ URLStore = (*Maps)(nil)

type Maps struct {
	urlMap map[string]string
}

func NewStorage() *Maps {
	URLMap := make(map[string]string)
	return &Maps{
		urlMap: URLMap,
	}
}

func (mp *Maps) Get(short string) (string, error) {
	if short == "" {
		return "", errors.New("err short url")

	}
	longURL, ok := mp.urlMap[short]
	if ok {
		return longURL, nil
	}
	return "", nil
}
func (mp *Maps) Save(searchURL string) (string, error) {
	if searchURL == "" {
		return "", errors.New("err")

	}
	short, err := mp.genShort(searchURL, len(mp.urlMap))
	if err != nil {
		return "", err
	}
	mp.urlMap[short] = searchURL
	return short, nil
}

func (mp *Maps) genShort(searchURL string, Count int) (string, error) {
	short := handlers.GenerateLink(searchURL, strconv.Itoa(Count))
	_, ok := mp.urlMap[short]
	if ok {
		Count++
		shortURL, err := mp.genShort(searchURL, Count)
		if err != nil || (Count-len(mp.urlMap) > 10) {
			return "", errors.New("err generate")
		}
		return shortURL, nil
	}
	return short, nil
}
