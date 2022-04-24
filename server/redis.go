package server

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Astemirdum/user-app/server/models"
	"github.com/go-redis/redis/v8"
)

const (
	cacheTTL = time.Minute
)

type Cache struct {
	Client *redis.Client
}

func NewRedisClient(addr, passwd string) (*Cache, error) {
	client := redis.NewClient(
		&redis.Options{Addr: addr,
			Password: passwd},
	)
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}
	return &Cache{client}, nil
}

func (c *Cache) GetCache(ctx context.Context, key string) ([]*models.User, error) {
	res := c.Client.Get(ctx, key)
	if err := res.Err(); err != nil {
		return nil, err
	}
	data, err := res.Bytes()
	if err != nil {
		return nil, err
	}
	users := make([]*models.User, 0)
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (c *Cache) DeleteCache(ctx context.Context, key string) error {
	if err := c.Client.Del(ctx, key).Err(); err != nil {
		return err
	}
	return nil
}

func (c *Cache) SetCache(ctx context.Context, key string, users []*models.User) error {
	data, err := json.Marshal(users)
	if err != nil {
		return err
	}
	res := c.Client.Set(ctx, key, data, cacheTTL)
	if err := res.Err(); err != nil {
		return err
	}
	return nil
}
