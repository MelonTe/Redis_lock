// 该包使用redis实现了分布式锁，分布式锁内部封装了go-redis框架的redis-client
// 用于后续的请求交互
package redis_lock

import (
	"context"
	"errors"
	"redis_lock/utils"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrLockAcquiredByOthers = errors.New("lock is acquired by others")

func IsRetryableErr(err error) bool {
	return errors.Is(err, ErrLockAcquiredByOthers)
}

// 基于redis实现的分布式锁结构，保证了对称性
type RedisLock struct {
	LockOptions
	key    string
	token  string
	client *redis.Client
}

func NewRedisLock(key string, client *redis.Client, opts ...LockOption) *RedisLock {
	r := RedisLock{
		key:    key,
		token:  utils.GetProcessAndGoroutineIDStr(),
		client: client,
	}
	for _, opt := range opts {
		opt(&r.LockOptions)
	}

	repairLock(&r.LockOptions)
	return &r
}

// key不存在时，才能设置成功，set时携带上超时时间，单位为秒
func (rl *RedisLock) SetNEX(ctx context.Context, key, value string) (bool, error) {
	if key == "" || value == "" {
		return false, errors.New("redis SET keyNX or value can't be empty")
	}
	//测试连接
	_, err := rl.client.Ping(ctx).Result()
	if err != nil {
		return false, err
	}

	//设置值
	duration := time.Duration(rl.LockOptions.expireSeconds) * time.Second
	ok, err := rl.client.SetNX(ctx, key, value, duration).Result()
	if err != nil {
		return false, err
	}
	//锁被他人占有
	if !ok {
		return false, nil
	}
	return true, nil
}
