package psql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/Xrefullx/golang-shorturl/internal/storage"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

var _ storage.Storage = (*Storage)(nil)

// Storage implements Storage interface, provides storing data in psql database.
type Storage struct {
	shortURLRepo *shortURLRepository
	userRepo     *userRepository
	db           *sql.DB
	conStringDSN string

	wg                *sync.WaitGroup // wait group for async tasks
	storageAsyncEnded chan struct{}
	urlAsyncEnded     chan struct{}
	waitAsyncEnd      bool
}

// NewStorage inits new connection to psql storage.
func NewStorage(ctx context.Context, asyncEndedChan chan struct{}, conStringDSN string) (*Storage, error) {
	if conStringDSN == "" {
		return nil, fmt.Errorf("ошибка инициализации бд:%v", "строка соединения с бд пуста")
	}

	db, err := sql.Open("postgres", conStringDSN)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}
	if err := initBase(db); err != nil {
		return nil, err
	}

	st := &Storage{
		db:           db,
		conStringDSN: conStringDSN,
	}

	st.shortURLRepo = newShortURLRepository(ctx, db, st.urlAsyncEnded)
	st.userRepo = newUserRepository(db)

	st.runWaiterAsyncEnded()
	return st, nil
}

// WaitAsyncTasksEnded returns true if on shutting down we must wait async tasks ended
func (s *Storage) WaitAsyncTasksEnded() bool {
	return s.waitAsyncEnd
}

func (s *Storage) runWaiterAsyncEnded() {
	go func() {
		//  wait url repository async tasks ended
		<-s.urlAsyncEnded

		fmt.Println("storage async tasks ended")

		s.storageAsyncEnded <- struct{}{}
	}()
}

// URL returns urls repository.
func (s *Storage) URL() storage.URLRepository {
	if s.shortURLRepo != nil {
		return s.shortURLRepo
	}
	return s.shortURLRepo
}

// User returns users repository.
func (s *Storage) User() storage.UserRepository {
	return s.userRepo
}

// Ping checks database connection.
func (s *Storage) Ping() error {
	if s == nil || s.db == nil {
		return errors.New("db not initialized")
	}

	if err := s.db.Ping(); err != nil {
		return fmt.Errorf("ping for DSN (%s) failed: %w", s.conStringDSN, err)
	}

	return nil
}

// Close  closes database connection.
func (s Storage) Close() {
	if s.db == nil {
		return
	}

	s.db.Close()
	s.db = nil
}

// initBase drops all and inits database tables.
func initBase(db *sql.DB) error {
	row := db.QueryRow("DROP SCHEMA public CASCADE;CREATE SCHEMA public;")
	if row.Err() != nil {
		return row.Err()
	}
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS users (" +
		"		id uuid not null,						" +
		"		primary key (id));" +
		"	CREATE TABLE IF NOT EXISTS urls(" +
		"		id uuid not null," +
		"		user_id uuid not null," +
		"		srcurl varchar(2050) not null," +
		"		shorturl varchar (16) not null," +
		"		isdeleted boolean not null," +
		"		unique (shorturl)," +
		"		unique (srcurl)," +
		"		primary key (id)," +
		"		foreign key (user_id) references users (id)" +
		"	);")
	if row.Err() != nil {
		return err
	}
	return nil

}
