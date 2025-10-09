package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
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
	"webserver/internal/service/user"

	user_cache "webserver/internal/service/storage/redis/user"

	"github.com/ex-rate/logger"
	"github.com/sirupsen/logrus"
)

func main() {
	// Создаем контекст для обработки сигналов завершения
	notifyCtx, notify := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer notify()

	configPath := flag.String("config", "./config.yaml", "path to config file")

	flag.Parse()

	cfg := start(config.LoadConfig(*configPath))

	lvl := start(logrus.ParseLevel(cfg.Logger.Level))

	loggerConfig := logger.Config{
		Level:  lvl,
		Output: logger.OutputType(cfg.Logger.Output),
		Format: cfg.Logger.Format,
	}

	log := start(logger.New(loggerConfig))

	log.Info("Logger initialized")

	logrus.WithField("level", log.GetLevel()).Info("set log level")

	butler := NewButler()

	logrus.WithFields(logrus.Fields{
		"version": butler.BuildInfo.Version,
		"commit":  butler.BuildInfo.GitCommit,
		"date":    butler.BuildInfo.BuildDate,
	}).Info("starting service")
	defer logrus.Info("shutdown")

	elasticClient := start(elasticsearch.New([]string{cfg.Storage.ElasticSearch.Address}))

	addr := formatPostgresAddr(cfg.Storage.Postgres)

	spaceRepo := start(space_db.New(addr, elasticClient))

	spaceCacheLog := log.WithService("space_cache")
	spaceCache := start(space_cache.New(notifyCtx, cfg.Storage.Redis.Address, spaceCacheLog))

	rabbitLog := log.WithService("rabbit")
	rabbit := start(worker.New(
		worker.WithAddress(cfg.Storage.RabbitMQ.Address),
		worker.WithNotesExchange(cfg.Storage.RabbitMQ.NoteExchange),
		worker.WithSpacesExchange(cfg.Storage.RabbitMQ.SpaceExchange),
		worker.WithLogger(rabbitLog),
	))
	go butler.start(rabbit.Connect)

	spaceSrvLog := log.WithService("space_srv")
	spaceSrv := start(space.New(
		space.WithRepo(spaceRepo),
		space.WithCache(spaceCache),
		space.WithWorker(rabbit),
		space.WithLogger(spaceSrvLog),
	))

	userCacheLog := log.WithService("user_cache")
	userCache := start(user_cache.New(notifyCtx, cfg.Storage.Redis.Address, userCacheLog))

	userRepo := start(user_db.New(addr))

	userSrvLog := log.WithService("user_srv")
	userSrv := start(user.New(
		user.WithRepo(userRepo),
		user.WithCache(userCache),
		user.WithLogger(userSrvLog),
	))

	authSrvLog := log.WithService("auth_srv")
	authSrv := start(auth.New(
		auth.WithSecretKey([]byte(cfg.Auth.SecretKey)),
		auth.WithLogger(authSrvLog),
	))

	handlerLog := log.WithService("handler")
	handler := start(v0.New(
		v0.WithSpaceService(spaceSrv),
		v0.WithUserService(userSrv),
		v0.WithAuthService(authSrv),
		v0.WithLogger(handlerLog),
		v0.WithVersion(butler.BuildInfo.Version),
		v0.WithBuildDate(butler.BuildInfo.BuildDate),
		v0.WithGitCommit(butler.BuildInfo.GitCommit),
	))

	serverLog := log.WithService("server")
	server := start(server.New(
		server.WithAddr(cfg.Server.Address),
		server.WithHandler(handler),
		server.WithLogger(serverLog),
	))

	startService(server.CreateRoutes(), "server routes")

	log.WithField("address", cfg.Server.Address).Info("started server")

	go butler.start(server.Start)

	log.Info("all services started")

	// Ждем сигнал завершения
	<-notifyCtx.Done()
	log.Info("received shutdown signal, stopping services...")

	// Контекст завершения с таймаутом для graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Останавливаем сервисы в обратном порядке и те, у которых есть фоновые горутины
	// 1) HTTP сервер (разбудит Start и освободит горутину)
	butler.stop(shutdownCtx, server)
	// 2) Rabbit worker (разбудит Connect горутину)
	butler.stop(shutdownCtx, rabbit)
	// 3) Доменные сервисы
	butler.stop(shutdownCtx, spaceSrv)
	butler.stop(shutdownCtx, userSrv)
	butler.stop(shutdownCtx, authSrv)
	// 4) Репозитории/кеши/клиенты
	butler.stop(shutdownCtx, spaceRepo)
	butler.stop(shutdownCtx, userRepo)
	butler.stop(shutdownCtx, spaceCache)
	butler.stop(shutdownCtx, userCache)
	butler.stop(shutdownCtx, elasticClient)

	// Ждем завершения всех горутин после остановки сервисов
	butler.waitForAll()
	log.Info("all services stopped")
}

func startService(err error, name string) {
	if err != nil {
		// Используем logrus для критических ошибок, так как наш логгер может быть еще не инициализирован
		logrus.Fatalf("error creating %s: %+v", name, err)
	}
}

func start[T any](svc T, err error) T {
	startService(err, fmt.Sprintf("%T", svc))

	return svc
}

func formatPostgresAddr(cfg config.Postgres) string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.User, cfg.Password,
		cfg.Host, cfg.Port, cfg.DBName)
}
