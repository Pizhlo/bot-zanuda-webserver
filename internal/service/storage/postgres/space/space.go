package space

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/sirupsen/logrus"
)

type spaceRepo struct {
	db        *sql.DB
	currentTx *sql.Tx
}

func New(addr string) (*spaceRepo, error) {
	db, err := sql.Open("postgres", addr)
	if err != nil {
		return nil, fmt.Errorf("connect open a db driver: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("cannot connect to a db: %w", err)
	}
	return &spaceRepo{db, nil}, nil
}

func (db *spaceRepo) Close() {
	if err := db.db.Close(); err != nil {
		logrus.Errorf("error on closing space repo: %v", err)
	}
}

func (db *spaceRepo) tx(ctx context.Context) (*sql.Tx, error) {
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

func (db *spaceRepo) commit() error {
	tx := db.currentTx
	db.currentTx = nil
	return tx.Commit()
}
