package space

import (
	"context"
	"fmt"
	"strconv"
	"time"
	"webserver/internal/model"

	api_errors "webserver/internal/errors"

	"github.com/ex-rate/logger"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
	logger *logger.Logger
}

func New(ctx context.Context, addr string, logger *logger.Logger) (*Cache, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	status := redisClient.Ping(ctx)

	if err := status.Err(); err != nil {
		return nil, err
	}

	logger.WithField("addr", addr).Info("successfully connected redis")

	return &Cache{
		client: redisClient,
		logger: logger,
	}, nil
}

const (
	// запрос получить пользователя по айди. пример: HGETALL user:297850813
	spaceKey = "space:%s"

	// ключи, хранящиеся в редисе
	idKey       = "id"
	nameKey     = "name"
	createdKey  = "created"
	personalKey = "personal"
	creatorKey  = "creator"
)

func (s *Cache) GetSpaceByID(ctx context.Context, id uuid.UUID) (model.Space, error) {
	key := fmt.Sprintf(spaceKey, id.String())
	res, err := s.client.HGetAll(ctx, key).Result()
	if err == redis.Nil {
		return model.Space{}, api_errors.ErrSpaceNotExists
	}

	if err != nil {
		return model.Space{}, err
	}

	s.logger.WithField("res", res).Debug("got space from redis")

	if len(res) == 0 {
		return model.Space{}, api_errors.ErrSpaceNotExists
	}

	return parseSpace(res)
}

func parseSpace(res map[string]string) (model.Space, error) {
	layout := "2006-01-2 15:04:05"
	created, err := time.Parse(layout, res[createdKey])
	if err != nil {
		return model.Space{}, fmt.Errorf("error converting string created '%s' to time: %+v", res[createdKey], err)
	}

	creator, err := strconv.Atoi(res[creatorKey])
	if err != nil {
		return model.Space{}, fmt.Errorf("error converting string tg id '%s' to int: %+v", res[createdKey], err)
	}

	personal, err := strconv.ParseBool(res[personalKey])
	if err != nil {
		return model.Space{}, fmt.Errorf("error converting string personal '%s' to bool: %+v", res[createdKey], err)
	}

	id, err := uuid.Parse(res[idKey])
	if err != nil {
		return model.Space{}, fmt.Errorf("error converting string space id '%s' to uuid: %+v", res[idKey], err)
	}

	return model.Space{
		ID:       id,
		Name:     res[nameKey],
		Created:  created,
		Creator:  int64(creator),
		Personal: personal,
	}, nil
}
