package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"webserver/internal/config"
	"webserver/internal/server"
	v0 "webserver/internal/server/api/v0"
	"webserver/internal/service/auth"
	"webserver/internal/service/space"
	"webserver/internal/service/storage/elasticsearch"
	space_db "webserver/internal/service/storage/postgres/space"
	user_db "webserver/internal/service/storage/postgres/user"
	"webserver/internal/service/storage/rabbit/worker"
	space_cache "webserver/internal/service/storage/redis/space"
	user_cache "webserver/internal/service/storage/redis/user"
	"webserver/internal/service/user"

	"github.com/sirupsen/logrus"
)

// @title           Веб-сервер для Бота Зануды
// @description     Веб-сервер, обрабатывающий запросы от Бота Зануды: управление заметками, а также перенаправление запросов к другим сервисам (напоминяний, пользователей)
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	configPath := flag.String("config", "internal/config/config.yaml", "путь к файлу конфигурации")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		logrus.Fatalf("error loading config: %+v", err)
	}

	switch cfg.LogLevel {
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
		logrus.SetLevel(logrus.FatalLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}

	logrus.Infof("log level: %+v", logrus.GetLevel())

	elasticClient := start(elasticsearch.New([]string{cfg.Storage.ElasticSearch.Address}))

	addr := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Storage.Postgres.User, cfg.Storage.Postgres.Password, cfg.Storage.Postgres.Host, cfg.Storage.Postgres.Port, cfg.Storage.Postgres.DBName)

	logrus.Infof("connecting db on %s", addr)
	spaceRepo := start(space_db.New(addr, elasticClient))

	spaceCache := start(space_cache.New(ctx, cfg.Storage.Redis.Address))

	logrus.Infof("connecting rabbit on %s", cfg.Storage.RabbitMQ.Address)

	rabbit := start(worker.New(
		worker.WithAddress(cfg.Storage.RabbitMQ.Address),
		worker.WithNotesTopic(cfg.Storage.RabbitMQ.NoteQueue),
		worker.WithSpacesTopic(cfg.Storage.RabbitMQ.SpaceQueue),
	))

	startService(rabbit.Connect(), "rabbit")

	logrus.Infof("succesfully connected rabbit on %s", cfg.Storage.RabbitMQ.Address)

	spaceSrv := start(space.New(
		space.WithRepo(spaceRepo),
		space.WithCache(spaceCache),
		space.WithWorker(rabbit),
	))

	userCache := start(user_cache.New(ctx, cfg.Storage.Redis.Address))

	userRepo := start(user_db.New(addr))

	userSrv := start(user.New(
		user.WithRepo(userRepo),
		user.WithCache(userCache),
	))

	authSrv := start(auth.New(
		auth.WithSecretKey([]byte(cfg.Auth.SecretKey)),
	))

	handler := start(v0.New(
		v0.WithSpaceService(spaceSrv),
		v0.WithUserService(userSrv),
		v0.WithAuthService(authSrv),
	))

	server := start(server.New(
		server.WithAddr(cfg.Server.Address),
		server.WithHandler(handler),
	))

	startService(server.CreateRoutes(), "server routes")

	logrus.Infof("started server on %s", cfg.Server.Address)

	startService(server.Start(), "server")

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

		err := server.Shutdown(ctx)
		if err != nil {
			logrus.Errorf("error shutdown server: %+v", err)
		}

		spaceRepo.Close()
		err = rabbit.Close()
		if err != nil {
			logrus.Errorf("error closing rabbit: %+v", err)
		}

		userRepo.Close()
	}(&wg)

	wg.Wait()

	notify()
}

func startService(err error, name string) {
	if err != nil {
		logrus.Fatalf("error creating %s: %+v", name, err)
	}
}

func start[T any](svc T, err error) T {
	startService(err, fmt.Sprintf("%T", svc))

	return svc
}
