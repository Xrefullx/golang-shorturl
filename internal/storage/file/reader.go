package file

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/Xrefullx/golang-shorturl/internal/storage/postgres/schema_postgres"
	"github.com/google/uuid"
	"os"
)

type fileReader struct {
	file    *os.File
	scanner *bufio.Scanner
}

func newFileReader(fileName string) (*fileReader, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, fmt.Errorf("ошибка инициализации чтения файла: %w", err)
	}

	return &fileReader{
		file:    file,
		scanner: bufio.NewScanner(file),
	}, nil
}

func (f *fileReader) Close() error {
	return f.file.Close()
}

func (f *fileReader) ReadAll() (map[uuid.UUID]schema_postgres.ShortURL, error) {
	data := make(map[uuid.UUID]schema_postgres.ShortURL)
	for f.scanner.Scan() {
		lineURL := schema_postgres.ShortURL{}
		if err := json.Unmarshal(f.scanner.Bytes(), &lineURL); err != nil {
			return nil, fmt.Errorf("ошибка обработки данных из файла: %w", err)
		}
		data[lineURL.ID] = lineURL
	}

	if err := f.scanner.Err(); err != nil {
		return nil, fmt.Errorf("ошибка чтения файла: %w", err)
	}

	return data, nil
}
