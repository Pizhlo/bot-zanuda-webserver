package app

import (
	"context"
	"fmt"
	"webserver/internal/config"
	"webserver/internal/server"
	v0 "webserver/internal/server/api/v0"
	auth "webserver/internal/service/auth"
	space "webserver/internal/service/space"
	space_db "webserver/internal/service/storage/postgres/space"
	user_db "webserver/internal/service/storage/postgres/user"
	worker "webserver/internal/service/storage/rabbit/worker"
	space_cache "webserver/internal/service/storage/redis/space"
	user_cache "webserver/internal/service/storage/redis/user"
	user "webserver/internal/service/user"

	"webserver/internal/service/storage/elasticsearch"

	"github.com/ex-rate/logger"
	"github.com/sirupsen/logrus"
)

type App struct {
	Cfg        *config.Config
	Elastic    *elasticsearch.Client
	SpaceRepo  *space_db.Repo
	SpaceCache *space_cache.Cache
	Rabbit     *worker.Worker
	SpaceSrv   *space.Service
	UserCache  *user_cache.Cache
	UserRepo   *user_db.Repo
	UserSrv    *user.Service
	AuthSrv    *auth.Service
	Handler    *v0.Handler
	Server     *server.Server
}

func NewApp(ctx context.Context, configPath string) (*App, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		logrus.Fatalf("error loading config: %+v", err)
	}

	lvl, err := logrus.ParseLevel(cfg.Logger.Level)
	if err != nil {
		logrus.Fatalf("error parsing level: %+v", err)
	}

	// Создание логгера
	loggerConfig := logger.Config{
		Level:  lvl,
		Output: logger.OutputType(cfg.Logger.Output),
		Format: cfg.Logger.Format,
	}

	log, err := logger.New(loggerConfig)
	if err != nil {
		logrus.Fatalf("error creating logger: %+v", err)
	}

	log.Info("Logger initialized")

	elasticClient := start(elasticsearch.New([]string{cfg.Storage.ElasticSearch.Address}))

	addr := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Storage.Postgres.User, cfg.Storage.Postgres.Password, cfg.Storage.Postgres.Host, cfg.Storage.Postgres.Port, cfg.Storage.Postgres.DBName)

	spaceRepo := start(space_db.New(addr, elasticClient))

	spaceCacheLog := log.WithService("space_cache")
	spaceCache := start(space_cache.New(ctx, cfg.Storage.Redis.Address, spaceCacheLog))

	rabbitLog := log.WithService("rabbit")
	rabbit := start(worker.New(
		worker.WithAddress(cfg.Storage.RabbitMQ.Address),
		worker.WithNotesExchange(cfg.Storage.RabbitMQ.NoteExchange),
		worker.WithSpacesExchange(cfg.Storage.RabbitMQ.SpaceExchange),
		worker.WithLogger(rabbitLog),
	))

	startService(rabbit.Connect(), "rabbit")

	spaceSrvLog := log.WithService("space_srv")
	spaceSrv := start(space.New(
		space.WithRepo(spaceRepo),
		space.WithCache(spaceCache),
		space.WithWorker(rabbit),
		space.WithLogger(spaceSrvLog),
	))

	userCacheLog := log.WithService("user_cache")
	userCache := start(user_cache.New(ctx, cfg.Storage.Redis.Address, userCacheLog))

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
	))

	serverLog := log.WithService("server")
	server := start(server.New(
		server.WithAddr(cfg.Server.Address),
		server.WithHandler(handler),
		server.WithLogger(serverLog),
	))

	startService(server.CreateRoutes(), "server routes")

	log.Infof("started server on %s", cfg.Server.Address)

	startService(server.Start(), "server")

	return &App{
		Cfg:        cfg,
		Elastic:    elasticClient,
		SpaceRepo:  spaceRepo,
		SpaceCache: spaceCache,
		Rabbit:     rabbit,
		SpaceSrv:   spaceSrv,
		UserCache:  userCache,
		UserRepo:   userRepo,
		UserSrv:    userSrv,
		AuthSrv:    authSrv,
		Handler:    handler,
		Server:     server,
	}, nil
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
