package auth

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
)

type UserStore struct {
	client *redis.Client
}

func NewUserStore(client *redis.Client) *UserStore {
	return &UserStore{client: client}
}

func (us *UserStore) CreateUser(ctx context.Context, username, password string) error {
	key := "user:" + username
	exists, err := us.client.Exists(ctx, key).Result()
	if err != nil {
		return err
	}
	if exists == 1 {
		return errors.New("user already exists")
	}
	return us.client.Set(ctx, key, password, 0).Err()
}

func (us *UserStore) ValidateUser(ctx context.Context, username, password string) (bool, error) {
	key := "user:" + username
	storedPassword, err := us.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return storedPassword == password, nil
}
