package user

import (
	"context"
	"fmt"
	"webserver/internal/model"

	"github.com/redis/go-redis/v9"
)

type userCache struct {
	client *redis.Client
}

func New(ctx context.Context, addr string) (*userCache, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	status := redisClient.Ping(ctx)

	if err := status.Err(); err != nil {
		return nil, err
	}

	return &userCache{
		client: redisClient,
	}, nil
}

func (u *userCache) Save(ctx context.Context, user model.User) error {
	return u.client.Set(ctx, fmt.Sprintf("%d", user.TgID), user, 0).Err()
}

func (u *userCache) GetUser(tgID int64) (model.User, error) {
	return model.User{}, nil
}
