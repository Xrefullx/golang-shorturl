package handlers

type URLStore interface {
	Get(shortURL string) (string, error)
	Save(searchURL string) (string, error)
}
