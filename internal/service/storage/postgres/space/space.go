package space

import (
	"context"
	"database/sql"
	"fmt"
	"webserver/internal/model/elastic"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	sqldblogger "github.com/simukti/sqldb-logger"
	"github.com/simukti/sqldb-logger/logadapter/logrusadapter"
	"github.com/sirupsen/logrus"
)

type spaceRepo struct {
	db            *sql.DB
	currentTx     *sql.Tx
	elasticClient elasticClient
}

//go:generate mockgen -source ./space.go -destination=../../../../../mocks/elastic.go -package=mocks
type elasticClient interface {
	Save(ctx context.Context, search elastic.Data) error
	// SearchByText производит поиск по тексту (названию). Возвращает ID из базы подходящих записей
	SearchByText(ctx context.Context, search elastic.Data) ([]uuid.UUID, error)
	// // SearchByID производит поиск по ID из базы. Возвращает ID  из эластика подходящих записей
	// SearchByID(ctx context.Context, search elastic.Data) ([]string, error)
	// Delete(ctx context.Context, search elastic.Data) error
	// DeleteAllByUserID(ctx context.Context, data elastic.Data) error
	UpdateNote(ctx context.Context, search elastic.Data) error
}

func New(addr string, elasticClient elasticClient) (*spaceRepo, error) {
	db, err := sql.Open("postgres", addr)
	if err != nil {
		return nil, fmt.Errorf("connect open a db driver: %w", err)
	}

	logger := logrus.New()
	logger.Level = logrus.DebugLevel           // miminum level
	logger.Formatter = &logrus.JSONFormatter{} // logrus automatically add time field

	db = sqldblogger.OpenDriver(addr, db.Driver(), logrusadapter.New(logger) /*, using_default_options*/) // db is STILL *sql.DB
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("cannot connect to a db: %w", err)
	} // to check connectivity and DSN correctness

	return &spaceRepo{db, nil, elasticClient}, nil
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
