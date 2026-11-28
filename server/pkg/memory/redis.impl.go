package memory

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// redisDefaultExpiration 与本地实现 cache.New(5*time.Minute, ...) 的默认过期一致。
const redisDefaultExpiration = 5 * time.Minute

// Redis 基于 go-redis 的缓存实现，语义对齐本地 Memory（含过期约定）。
type Redis struct {
	client *redis.Client
	prefix string
}

// NewRedis 创建并连通性校验（Ping）；不可达返回错误，由调用方降级到本地缓存。
func NewRedis(addr, password string, db int, prefix string) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}
	return &Redis{client: client, prefix: prefix}, nil
}

func (r *Redis) k(key string) string {
	return r.prefix + key
}

// exp 映射 go-cache 过期约定：0=默认(5min)，<0=永不过期，>0=指定时长。
func exp(duration time.Duration) time.Duration {
	if duration == 0 {
		return redisDefaultExpiration
	}
	if duration < 0 {
		return 0
	}
	return duration
}

func (r *Redis) ctx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 3*time.Second)
}

func (r *Redis) SetStruct(key string, data any, duration time.Duration) {
	b, err := json.Marshal(data)
	if err != nil {
		return
	}
	ctx, cancel := r.ctx()
	defer cancel()
	_ = r.client.Set(ctx, r.k(key), b, exp(duration)).Err()
}

func (r *Redis) GetStruct(key string, data any) error {
	ctx, cancel := r.ctx()
	defer cancel()
	b, err := r.client.Get(ctx, r.k(key)).Bytes()
	if errors.Is(err, redis.Nil) {
		return errors.New("key does not exist")
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(b, data)
}

func (r *Redis) SetString(key string, value string, duration time.Duration) {
	ctx, cancel := r.ctx()
	defer cancel()
	_ = r.client.Set(ctx, r.k(key), value, exp(duration)).Err()
}

func (r *Redis) GetString(key string) (value string) {
	ctx, cancel := r.ctx()
	defer cancel()
	v, err := r.client.Get(ctx, r.k(key)).Result()
	if err != nil {
		return ""
	}
	return v
}

func (r *Redis) Get(key string, data AnyType) error {
	ctx, cancel := r.ctx()
	defer cancel()
	b, err := r.client.Get(ctx, r.k(key)).Bytes()
	if errors.Is(err, redis.Nil) {
		return errors.New("key does not exist")
	}
	if err != nil {
		return err
	}
	return data.Unmarshal(b)
}

func (r *Redis) Set(key string, data AnyType, duration time.Duration) {
	ctx, cancel := r.ctx()
	defer cancel()
	_ = r.client.Set(ctx, r.k(key), data.Marshal(), exp(duration)).Err()
}

func (r *Redis) Del(key string) {
	ctx, cancel := r.ctx()
	defer cancel()
	_ = r.client.Del(ctx, r.k(key)).Err()
}

// Sync Redis 自身持久化，无需落盘。
func (r *Redis) Sync() {}
