package storage

type URLStorage struct {
	URLStorage URLStore
}

type URLStore interface {
	Get(shortURL string) (string, error)
	Save(searchURL string) (string, error)
}
