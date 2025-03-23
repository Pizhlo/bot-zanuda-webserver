package space

import (
	"context"
	"fmt"
	"strconv"
	"time"
	"webserver/internal/model"

	api_errors "webserver/internal/errors"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type spaceCache struct {
	client *redis.Client
}

func New(ctx context.Context, addr string) (*spaceCache, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	status := redisClient.Ping(ctx)

	if err := status.Err(); err != nil {
		return nil, err
	}

	logrus.Infof("space cache: successfully connected redis on %s", addr)

	return &spaceCache{
		client: redisClient,
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

func (s *spaceCache) GetSpaceByID(ctx context.Context, id uuid.UUID) (model.Space, error) {
	key := fmt.Sprintf(spaceKey, id.String())
	res, err := s.client.HGetAll(ctx, key).Result()
	if err == redis.Nil {
		return model.Space{}, api_errors.ErrSpaceNotExists
	}

	if err != nil {
		return model.Space{}, err
	}

	logrus.Debugf("got space from redis: %+v", res)

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
