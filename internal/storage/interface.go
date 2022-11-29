package storage

type URLStore interface {
	Get(short string) (string, error)
	Save(short string, search string) (string, error)
	IsShort(short string) bool
}
