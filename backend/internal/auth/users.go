package auth

import (
	"context"
	"errors"
	"regexp"

	"github.com/redis/go-redis/v9"
)

var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9]{4,32}$`)
	passwordRegex = regexp.MustCompile(`^[a-zA-Z0-9!@#$%^&*()_+\-=\[\]{}|;:,.<>?/]{4,32}$`)
)

type UserStore struct {
	client *redis.Client
}

func NewUserStore(client *redis.Client) *UserStore {
	return &UserStore{client: client}
}

func (us *UserStore) CreateUser(ctx context.Context, username, password string) error {
	if !usernameRegex.MatchString(username) {
		return errors.New("invalid username: only latin letters and digits, 4-32 chars")
	}
	if !passwordRegex.MatchString(password) {
		return errors.New("invalid password: only latin letters, digits and special chars (!@#$%^&*()_+-=[]{}|;:,.<>?/), 4-32 chars")
	}
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
	if !usernameRegex.MatchString(username) {
		return false, errors.New("invalid username: only latin letters and digits, 4-32 chars")
	}
	if !passwordRegex.MatchString(password) {
		return false, errors.New("invalid password: only latin letters, digits and special chars (!@#$%^&*()_+-=[]{}|;:,.<>?/), 4-32 chars")
	}
	key := "user:" + username
	storedPassword, err := us.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return storedPassword == password, nil
}
