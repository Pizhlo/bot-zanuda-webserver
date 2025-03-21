package space

import (
	"context"

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

	logrus.Infof("successfully connected redis on %s", addr)

	return &spaceCache{
		client: redisClient,
	}, nil
}

func (u *spaceCache) CheckParticipant(ctx context.Context, userID int64, spaceID uuid.UUID) error {
	return nil
}
