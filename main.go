package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"webserver/internal/app"

	"github.com/sirupsen/logrus"
)

// @title           Веб-сервер для Бота Зануды
// @description     Веб-сервер, обрабатывающий запросы от Бота Зануды: управление заметками, а также перенаправление запросов к другим сервисам (напоминяний, пользователей)
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	configPath := flag.String("config", "internal/config/config.yaml", "путь к файлу конфигурации")
	flag.Parse()

	app, err := app.NewApp(ctx, *configPath)
	if err != nil {
		logrus.Fatalf("error creating app: %+v", err)
	}

	notifyCtx, notify := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer notify()

	<-notifyCtx.Done()
	logrus.Info("shutdown")

	var wg sync.WaitGroup

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		logrus.Infof("shutdown server on %s", app.Cfg.Server.Address)

		ctx, cancel := context.WithTimeout(notifyCtx, app.Cfg.Server.ShutdownTimeout)
		defer cancel()

		err := app.Server.Shutdown(ctx)
		if err != nil {
			logrus.Errorf("error shutdown server: %+v", err)
		}

		app.SpaceRepo.Close()
		err = app.Rabbit.Close()
		if err != nil {
			logrus.Errorf("error closing rabbit: %+v", err)
		}

		app.UserRepo.Close()
	}(&wg)

	wg.Wait()

	notify()
}
