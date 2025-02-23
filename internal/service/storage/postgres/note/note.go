package note

import (
	"context"
	"database/sql"
	"fmt"
	"webserver/internal/model/elastic"

	"github.com/sirupsen/logrus"
)

type noteRepo struct {
	db            *sql.DB
	currentTx     *sql.Tx
	elasticClient elasticClient
}

//go:generate mockgen -source ./note.go -destination=../../../../../mocks/elastic.go -package=mocks
type elasticClient interface {
	Save(ctx context.Context, search elastic.Data) error
	// SearchByText производит поиск по тексту (названию). Возвращает ID из базы подходящих записей
	// SearchByText(ctx context.Context, search elastic.Data) ([]uuid.UUID, error)
	// // SearchByID производит поиск по ID из базы. Возвращает ID  из эластика подходящих записей
	// SearchByID(ctx context.Context, search elastic.Data) ([]string, error)
	// Delete(ctx context.Context, search elastic.Data) error
	// DeleteAllByUserID(ctx context.Context, data elastic.Data) error
	// Update(ctx context.Context, search elastic.Data) error
}

func New(addr string, elasticClient elasticClient) (*noteRepo, error) {
	db, err := sql.Open("postgres", addr)
	if err != nil {
		return nil, fmt.Errorf("connect open a db driver: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("cannot connect to a db: %w", err)
	}
	return &noteRepo{db, nil, elasticClient}, nil
}

func (db *noteRepo) Close() {
	if err := db.db.Close(); err != nil {
		logrus.Errorf("error on closing note repo: %v", err)
	}
}

func (db *noteRepo) tx(ctx context.Context) (*sql.Tx, error) {
	if db.currentTx != nil {
		return db.currentTx, nil
	}

	tx, err := db.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	})
	if err != nil {
		return nil, err
	}

	db.currentTx = tx

	return tx, nil
}

func (db *noteRepo) commit() error {
	tx := db.currentTx
	db.currentTx = nil
	return tx.Commit()
}

// func (db *noteRepo) rollback() error {
// 	tx := db.currentTx
// 	db.currentTx = nil
// 	return tx.Rollback()
// }
