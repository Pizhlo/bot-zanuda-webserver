package user

import (
	"context"
	"fmt"
	"strconv"
	"webserver/internal/model"

	api_errors "webserver/internal/errors"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type Cache struct {
	client *redis.Client
}

func New(ctx context.Context, addr string) (*Cache, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	status := redisClient.Ping(ctx)

	if err := status.Err(); err != nil {
		return nil, err
	}

	logrus.Infof("user cache: successfully connected redis on %s", addr)

	return &Cache{
		client: redisClient,
	}, nil
}

func (u *Cache) Save(ctx context.Context, user model.User) error {
	return u.client.Set(ctx, fmt.Sprintf("%d", user.TgID), user, 0).Err()
}

const (
	// запрос получить пользователя по айди. пример: HGETALL user:297850813
	userKey = "user:%d"

	// ключи, хранящиеся в редисе
	personalSpaceIDKey = "personal_space_id"
	telegramIDKey      = "telegram_id"
	timezoneKey        = "timezone"
	usernameKey        = "username"
	idKey              = "id"
)

func (u *Cache) GetUser(ctx context.Context, tgID int64) (model.User, error) {
	res, err := u.client.HGetAll(ctx, fmt.Sprintf(userKey, tgID)).Result()
	if err == redis.Nil {
		return model.User{}, api_errors.ErrUnknownUser
	}

	if err != nil {
		return model.User{}, err
	}

	logrus.Debugf("got user from redis: %+v", res)

	if len(res) == 0 {
		return model.User{}, api_errors.ErrUnknownUser
	}

	return parseUser(res)
}

func (u *Cache) CheckUser(ctx context.Context, tgID int64) (bool, error) {
	user, err := u.GetUser(ctx, tgID)
	if err != nil {
		return false, err
	}

	return user.ID != 0, nil
}

func parseUser(res map[string]string) (model.User, error) {
	id, err := strconv.Atoi(res[idKey])
	if err != nil {
		return model.User{}, fmt.Errorf("error converting string id '%s' to int: %+v", res[idKey], err)
	}

	tgID, err := strconv.Atoi(res[telegramIDKey])
	if err != nil {
		return model.User{}, fmt.Errorf("error converting string tg id '%s' to int: %+v", res[telegramIDKey], err)
	}

	personalSpaceID, err := uuid.Parse(res[personalSpaceIDKey])
	if err != nil {
		return model.User{}, fmt.Errorf("error converting string personal space id '%s' to uuid: %+v", res[personalSpaceIDKey], err)
	}

	return model.User{
		ID:       id,
		TgID:     int64(tgID),
		Username: res[usernameKey],
		PersonalSpace: &model.Space{
			ID: personalSpaceID,
		},
		Timezone: res[timezoneKey],
	}, nil
}
