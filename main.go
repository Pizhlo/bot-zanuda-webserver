package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"webserver/internal/server"
	"webserver/internal/service/note"
	"webserver/internal/service/space"
	note_db "webserver/internal/service/storage/postgres/note"
	space_db "webserver/internal/service/storage/postgres/space"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// @title           Веб-сервер для Бота Зануды
// @description     Веб-сервер, обрабатывающий запросы от Бота Зануды: управление заметками, а также перенаправление запросов к другим сервисам (напоминяний, пользователей)
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := godotenv.Load()
	if err != nil {
		logrus.Errorf("error loading env: %+v", err)
	}

	logLvl := os.Getenv("LOG_LEVEL")

	switch logLvl {
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	case "trace":
		logrus.SetLevel(logrus.TraceLevel)
	case "panic":
		logrus.SetLevel(logrus.PanicLevel)
	case "fatal":
		logrus.SetLevel(logrus.PanicLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}

	logrus.Infof("log level: %+v", logrus.GetLevel())

	dbUser := os.Getenv("POSTGRES_USER")
	if len(dbUser) == 0 {
		logrus.Fatal("POSTGRES_USER is not set")
	}

	dbPass := os.Getenv("POSTGRES_PASSWORD")
	if len(dbPass) == 0 {
		logrus.Fatal("POSTGRES_PASSWORD is not set")
	}

	dbName := os.Getenv("POSTGRES_DB")
	if len(dbName) == 0 {
		logrus.Fatal("POSTGRES_DB is not set")
	}

	dbHost := os.Getenv("POSTGRES_HOST")
	if len(dbHost) == 0 {
		logrus.Fatal("POSTGRES_HOST is not set")
	}

	dbPort := os.Getenv("POSTGRES_PORT")
	if len(dbPort) == 0 {
		logrus.Fatal("POSTGRES_PORT is not set")
	}

	addr := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbName)

	logrus.Infof("connecting db on %s", addr)
	noteRepo, err := note_db.New(addr)
	if err != nil {
		logrus.Fatalf("error connecting db: %+v", err)
	}

	noteSrv := note.New(noteRepo)

	spaceRepo, err := space_db.New(addr)
	if err != nil {
		logrus.Fatalf("error connecting db: %+v", err)
	}

	spaceSrv := space.New(spaceRepo)

	serverAddr := os.Getenv("SERVER_ADDR")
	if len(serverAddr) == 0 {
		logrus.Fatalf("SERVER_ADDR not set")
	}

	logrus.Infof("starting server on %s", serverAddr)
	s := server.New(serverAddr, noteSrv, spaceSrv)

	err = s.Serve()
	if err != nil {
		logrus.Fatalf("error starting server: %+v", err)
	}

	logrus.Infof("started server on %s", serverAddr)

	notifyCtx, notify := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer notify()

	<-notifyCtx.Done()
	logrus.Info("shutdown")

	var wg sync.WaitGroup

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		ctx, cancel := context.WithTimeout(notifyCtx, 2*time.Second)
		defer cancel()

		err := s.Shutdown(ctx)
		if err != nil {
			logrus.Errorf("error shutdown server: %+v", err)
		}

		noteRepo.Close()

		spaceRepo.Close()
	}(&wg)

	wg.Wait()

	notify()
}
