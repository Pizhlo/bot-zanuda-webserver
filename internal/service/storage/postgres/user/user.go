package user

import (
	"database/sql"
	"fmt"

	sqldblogger "github.com/simukti/sqldb-logger"
	"github.com/simukti/sqldb-logger/logadapter/logrusadapter"
	"github.com/sirupsen/logrus"
)

type userRepo struct {
	db        *sql.DB
	currentTx *sql.Tx
}

func New(addr string) (*userRepo, error) {
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

	return &userRepo{db, nil}, nil
}

func (db *userRepo) Close() {
	if err := db.db.Close(); err != nil {
		logrus.Errorf("error on closing space repo: %v", err)
	}
}

// func (db *userRepo) tx(ctx context.Context) (*sql.Tx, error) {
// 	if db.currentTx != nil {
// 		return db.currentTx, nil
// 	}

// 	tx, err := db.db.BeginTx(ctx, &sql.TxOptions{
// 		Isolation: sql.LevelReadCommitted,
// 		ReadOnly:  false,
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	db.currentTx = tx

// 	return tx, nil
// }

// func (db *userRepo) commit() error {
// 	tx := db.currentTx
// 	db.currentTx = nil
// 	return tx.Commit()
// }
