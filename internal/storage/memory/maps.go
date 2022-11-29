package memory

import (
	"errors"
	"github.com/Xrefullx/golang-shorturl/internal/app"
	"github.com/Xrefullx/golang-shorturl/internal/storage"
	"strconv"
	"sync"
)

var _ storage.URLStore = (*Maps)(nil)

type Maps struct {
	urlMap map[string]string
	mutex  *sync.RWMutex
}

func NewStorage() *Maps {
	return &Maps{
		urlMap: make(map[string]string),
		mutex:  &sync.RWMutex{},
	}
}

func (mp *Maps) Get(short string) (string, error) {
	if short == "" {
		return "", errors.New("err short url")

	}
	mp.mutex.RLock()
	long, ok := mp.urlMap[short]
	defer mp.mutex.RUnlock()
	if ok {
		return long, nil
	}

	return "", nil
}
func (mp *Maps) Save(short string, search string) (string, error) {
	if search == "" {
		return "", errors.New("err")

	}
	short, err := mp.genShort(search, len(mp.urlMap))
	if err != nil {
		return "", err
	}
	mp.urlMap[short] = search

	mp.mutex.Lock()
	mp.urlMap[short] = search
	defer mp.mutex.Unlock()

	return short, nil
}

func (mp *Maps) genShort(searchURL string, Count int) (string, error) {
	short := app.GenerateLink(searchURL, strconv.Itoa(Count))
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
func (mp *Maps) IsShort(shortID string) bool {
	_, ok := mp.urlMap[shortID]

	return !ok
}
