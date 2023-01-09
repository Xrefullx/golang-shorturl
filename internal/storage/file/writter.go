package file

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/Xrefullx/golang-shorturl/internal/storage/postgres/schema_postgres"
	"os"
)

type fileWriter struct {
	file   *os.File
	writer *bufio.Writer
}

func newFileWriter(filename string) (*fileWriter, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return nil, fmt.Errorf("ошибка инициализации записи в файл: %w", err)
	}

	return &fileWriter{
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}

func (f *fileWriter) Close() error {
	return f.file.Close()
}

func (f *fileWriter) WriteURL(sht schema_postgres.ShortURL) error {

	jsURL, err := json.Marshal(sht)
	if err != nil {
		return fmt.Errorf("ошибка обработки данных для записи в файл: %w", err)
	}

	if _, err := f.writer.Write(jsURL); err != nil {
		return fmt.Errorf("ошибка записи в файл: %w", err)
	}
	if err := f.writer.WriteByte('\n'); err != nil {
		return fmt.Errorf("ошибка записи в файл: %w", err)
	}

	return f.writer.Flush()
}
