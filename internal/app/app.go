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

	logrus.Infof("connecting redis on %s", cfg.Storage.Redis.Address)
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
		logrus.Fatalf("error creating %s: %+v", name, err)
	}
}

func start[T any](svc T, err error) T {
	startService(err, fmt.Sprintf("%T", svc))

	return svc
}
