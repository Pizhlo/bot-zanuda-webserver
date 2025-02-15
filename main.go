package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"webserver/internal/server"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

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

	serverAddr := os.Getenv("SERVER_ADDR")
	if len(serverAddr) == 0 {
		logrus.Fatalf("SERVER_ADDR not set")
	}

	logrus.Infof("starting server on %s", serverAddr)
	s := server.New(serverAddr)

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
	}(&wg)

	wg.Wait()

	notify()
}
