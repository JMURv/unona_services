package redis

import (
	"context"
	"encoding/json"
	errs "github.com/JMURv/unona/services/internal/cache"
	cfg "github.com/JMURv/unona/services/pkg/config"
	"github.com/JMURv/unona/services/pkg/model"
	"github.com/go-redis/redis/v8"
	"github.com/opentracing/opentracing-go"
	"log"
	"time"
)

type Cache struct {
	cli *redis.Client
}

func New(conf *cfg.RedisConfig) *Cache {
	redisCli := redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Pass,
		DB:       0,
	})
	_, err := redisCli.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	return &Cache{cli: redisCli}
}

func (c *Cache) Close() {
	if err := c.cli.Close(); err != nil {
		log.Println("Failed to close connection to Redis: ", err)
	}
}

func (c *Cache) Get(ctx context.Context, key string) (*model.Notification, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "media.GetFromCache")
	defer span.Finish()

	val, err := c.cli.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, errs.ErrNotFoundInCache
	} else if err != nil {
		return nil, err
	}

	rev := &model.Notification{}
	if err = json.Unmarshal(val, rev); err != nil {
		return nil, err
	}
	return rev, nil
}

func (c *Cache) Set(ctx context.Context, t time.Duration, key string, r *model.Notification) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "media.SetToCache")
	defer span.Finish()

	bytes, err := json.Marshal(r)
	if err != nil {
		return err
	}

	if err = c.cli.Set(ctx, key, bytes, t).Err(); err != nil {
		return err
	}
	return nil
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "media.DeleteFromCache")
	defer span.Finish()

	if err := c.cli.Del(ctx, key).Err(); err != nil {
		return err
	}
	return nil
}
