package repo

import (
	"context"
	"time"

	redis "github.com/redis/go-redis/v9"
)

type RedisClientInterface interface {
	Set(ctx context.Context, key string, value interface{}, expire time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key ...string) error
	Pub(ctx context.Context, channel string, message interface{}) error
	Sub(ctx context.Context, channel string) (<-chan string, error)
}

type redisRepo struct {
	Client redis.Cmdable
}

func NewRedisRepo(Client redis.Cmdable) RedisClientInterface {
	return &redisRepo{Client: Client}
}

func (r *redisRepo) Set(ctx context.Context, key string, value interface{}, expire time.Duration) error {
	return r.Client.Set(ctx, key, value, expire).Err()
}

func (r *redisRepo) Get(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}
func (r *redisRepo) Del(ctx context.Context, key ...string) error {
	return r.Client.Del(ctx, key...).Err()
}
func (r *redisRepo) Pub(ctx context.Context, channel string, message interface{}) error {
	return r.Client.Publish(ctx, channel, message).Err()
}

// idk what this function does...
func (r *redisRepo) Sub(ctx context.Context, channel string) (<-chan string, error) {
	pubsub := r.Client.(*redis.Client).Subscribe(ctx, channel)
	ch := make(chan string)
	go func() {
		defer close(ch)
		defer pubsub.Close()
		for msg := range pubsub.Channel() {
			ch <- msg.Payload
		}
	}()
	return ch, nil
}
