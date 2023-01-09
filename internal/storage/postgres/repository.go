package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Xrefullx/golang-shorturl/internal/model"
	"github.com/Xrefullx/golang-shorturl/internal/shrterr"
	st "github.com/Xrefullx/golang-shorturl/internal/storage"
	"github.com/Xrefullx/golang-shorturl/internal/storage/postgres/schema_postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	"log"
	"sync"
)

var _ st.URLRepository = (*shortURLRepository)(nil)

type URLBuffer struct {
	buf []model.ShortURL
	sync.Mutex
}

type shortURLRepository struct {
	db              *sql.DB
	insertBuffer    URLBuffer
	deleteChan      chan schema_postgres.ShortURL
	flushDeleteChan chan struct{}
}

const (
	delBuffBatch = 10
)

func newShortURLRepository(db *sql.DB) *shortURLRepository {
	repo := shortURLRepository{
		db: db,
		insertBuffer: URLBuffer{
			buf: make([]model.ShortURL, 0, 100),
		},
		deleteChan:      make(chan schema_postgres.ShortURL, delBuffBatch),
		flushDeleteChan: make(chan struct{}),
	}
	repo.initDeleteBatchWorker()

	return &repo
}

func (r *shortURLRepository) DeleteURLBatch(userID uuid.UUID, shortIDList ...string) error {
	if len(shortIDList) == 0 {
		return nil
	}

	go func() {
		for _, v := range shortIDList {
			r.deleteChan <- schema_postgres.ShortURL{ShortID: v, UserID: userID}
		}

		//run flush on end of list
		r.flushDeleteChan <- struct{}{}
	}()

	return nil
}

// Init single delete worker, that takes URLs from deleteChan and delete when filling the cache,
// or when take signal from flushDeleteChan
func (r *shortURLRepository) initDeleteBatchWorker() {
	go func() {
		cache := make([]schema_postgres.ShortURL, 0, delBuffBatch)
		for {
			select {
			// read URL to delete from deleteChan
			case v := <-r.deleteChan:
				cache = append(cache, v)
				if len(cache) < cap(cache) {
					continue
				}
			// read flush signal
			case <-r.flushDeleteChan:
				if len(cache) == 0 {
					continue
				}
			}

			if err := r.deleteTxURLBatch(cache); err != nil {
				log.Fatalf("ошибка транзакции удаления очереди URL:%v", err.Error())
			}
			cache = make([]schema_postgres.ShortURL, 0, delBuffBatch)
		}
	}()
}

// Delete array of URLs with transaction
func (r *shortURLRepository) deleteTxURLBatch(urls []schema_postgres.ShortURL) (err error) {
	tx, err := r.db.Begin()
	if err != nil {
		return
	}

	// defer make rollback
	defer func() {
		if err != nil {
			if rollErr := tx.Rollback(); rollErr != nil {
				err = fmt.Errorf("ошибка транзакции удаления:%v; транзакцию не удалось отменить:%w", err.Error(), rollErr)
			}
		}
	}()

	stmt, err := tx.Prepare("UPDATE urls SET isdeleted = TRUE WHERE shorturl = $1 AND user_id = $2")
	if err != nil {
		return
	}

	for _, sht := range urls {
		_, err = stmt.Exec(sht.ShortID, sht.UserID)
		if err != nil {
			return
		}
	}

	if err = tx.Commit(); err != nil {
		return
	}

	return nil
}

func (r *shortURLRepository) SaveURLBuff(sht *model.ShortURL) error {
	if sht == nil {
		return errors.New("URL is nil")
	}

	r.insertBuffer.Lock()
	defer r.insertBuffer.Unlock()

	r.insertBuffer.buf = append(r.insertBuffer.buf, *sht)

	if cap(r.insertBuffer.buf) == len(r.insertBuffer.buf) {
		err := r.saveURLBuffFlushNoLock()
		if err != nil {
			return fmt.Errorf("ошибка хранилица:%w", err)
		}
	}
	return nil
}

func (r *shortURLRepository) SaveURLBuffFlush() error {
	r.insertBuffer.Lock()
	defer r.insertBuffer.Unlock()

	return r.saveURLBuffFlushNoLock()
}

func (r *shortURLRepository) saveURLBuffFlushNoLock() (err error) {
	tx, err := r.db.Begin()
	if err != nil {
		return
	}

	// defer make rollback and clean buffer
	defer func() {
		r.insertBuffer.buf = r.insertBuffer.buf[:0]
		if err != nil {
			if rollErr := tx.Rollback(); rollErr != nil {
				err = fmt.Errorf("ошибка транзакции сохранения:%v; транзакцию не удалось отменить:%w", err.Error(), rollErr)
			}
		}
	}()

	stmt, err := tx.Prepare("INSERT INTO urls(id, user_id, srcurl, shorturl, isdeleted) VALUES($1, $2, $3, $4, $5)RETURNING id")
	if err != nil {
		return
	}

	for _, sht := range r.insertBuffer.buf {
		var dbObj schema_postgres.ShortURL
		dbObj, err = schema_postgres.NewURLFromCanonical(sht)
		if err != nil {
			continue
		}

		if err = stmt.QueryRow(
			dbObj.ID,
			dbObj.UserID,
			dbObj.URL,
			dbObj.ShortID,
			dbObj.IsDeleted).Scan(&dbObj.ID); err != nil {
			err = fmt.Errorf("ошибка транзакции сохранения dbObj- %v :%w ", dbObj.UserID.String(), err)

			return
		}
		sht.ID = dbObj.ID
	}

	if err = tx.Commit(); err != nil {
		return
	}

	return nil
}

func (r *shortURLRepository) SaveURL(ctx context.Context, sht model.ShortURL) (model.ShortURL, error) {
	dbObj, err := schema_postgres.NewURLFromCanonical(sht)
	if err != nil {
		return model.ShortURL{}, fmt.Errorf("ошибка хранилица:%w", err)
	}

	row := r.db.QueryRowContext(
		ctx,
		"INSERT INTO urls (id, user_id, srcurl, shorturl, isdeleted) VALUES ($1, $2, $3, $4, $5) RETURNING id ",
		dbObj.ID,
		dbObj.UserID,
		dbObj.URL,
		dbObj.ShortID,
		dbObj.IsDeleted,
	)

	if row.Err() != nil {
		// check duplicate srcurl
		pqErr, ok := row.Err().(*pq.Error)
		if ok && pqErr.Code == pgerrcode.UniqueViolation && pqErr.Constraint == "urls_srcurl_key" {
			existURL, err := r.GetShortURLBySrcURL(ctx, sht.URL)
			if err != nil {
				return model.ShortURL{}, fmt.Errorf("ошибка добавления записи в БД, ссылка %v уже существует: ошибка получения существующей короткой ссыки: %w",
					sht.URL, err)
			}
			return model.ShortURL{}, &shrterr.ErrorConflictSaveURL{
				Err:           row.Err(),
				ExistShortURL: existURL.ShortID,
			}
		}
	}

	if err := row.Scan(&dbObj.ID); err != nil {
		return model.ShortURL{}, err
	}

	sht.ID = dbObj.ID
	return sht, nil
}

func (r *shortURLRepository) GetURL(ctx context.Context, shortID string) (model.ShortURL, error) {
	dbObj := schema_postgres.ShortURL{}
	err := r.db.QueryRow(
		"select id, user_id, srcurl, shorturl, isdeleted from urls where shorturl = $1", shortID,
	).Scan(&dbObj.ID, &dbObj.UserID, &dbObj.URL, &dbObj.ShortID, &dbObj.IsDeleted)

	if err != nil {
		return model.ShortURL{}, fmt.Errorf("ошибка хранилица:%w", err)
	}

	return dbObj.ToCanonical()
}

func (r *shortURLRepository) GetShortURLBySrcURL(ctx context.Context, url string) (model.ShortURL, error) {
	dbObj := schema_postgres.ShortURL{}
	err := r.db.QueryRow(
		"select id, user_id, srcurl, shorturl, isdeleted from urls where srcurl = $1", url,
	).Scan(&dbObj.ID, &dbObj.UserID, &dbObj.URL, &dbObj.ShortID, &dbObj.IsDeleted)

	if err != nil {
		return model.ShortURL{}, fmt.Errorf("ошибка хранилица:%w", err)
	}

	return dbObj.ToCanonical()
}

func (r *shortURLRepository) GetUserURLList(ctx context.Context, userID uuid.UUID, limit int) ([]model.ShortURL, error) {
	var userURLs schema_postgres.URLList
	userURLs = make([]schema_postgres.ShortURL, 0, limit)

	rows, err := r.db.QueryContext(
		ctx,
		"SELECT id, user_id, srcurl, shorturl, isdeleted from urls WHERE user_id = $1 LIMIT $2", userID, limit)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var s schema_postgres.ShortURL
		err = rows.Scan(&s.ID, &s.UserID, &s.URL, &s.ShortID, &s.IsDeleted)
		if err != nil {
			return nil, err
		}

		userURLs = append(userURLs, s)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return userURLs.ToCanonical()
}

func (r *shortURLRepository) Exist(shortID string) (bool, error) {
	count := 0
	err := r.db.QueryRow(
		"SELECT  COUNT(*) as count FROM urls WHERE shorturl = $1", shortID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
