package redis_lock

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

const RedisLockKeyPrefix = "REDIS_LOCK_PREFIX_"

// LOCK加锁
func (r *RedisLock) Lock(ctx context.Context) error {
	//不管是否阻塞，都需要先获取一次锁
	err := r.tryLock(ctx)
	if err == nil {
		return nil
	}

	//非阻塞模式加锁失败直接返回
	if !r.isBlock {
		return err
	}

	//基于阻塞模式持续轮询取锁
	return r.blockingLock(ctx)
}
func (r *RedisLock) tryLock(ctx context.Context) error {
	ok, err := r.SetNEX(ctx, r.getLockKey(), r.token)
	if err != nil {
		return err
	}
	if !ok {
		return ErrLockAcquiredByOthers
	}
	return nil
}
func (r *RedisLock) getLockKey() string {
	return RedisLockKeyPrefix + r.key
}

// 阻塞模式加锁
func (r *RedisLock) blockingLock(ctx context.Context) error {
	//阻塞模式等待锁时间上限
	timeoutCh := time.After(time.Duration(r.blockWaitingSeconds) * time.Second)
	//轮询ticker，每隔50ms尝试获取一次锁
	ticker := time.NewTicker(time.Duration(50) * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		select {
		//ctx终止了
		case <-ctx.Done():
			return fmt.Errorf("lock failed, ctx timeout, err: %w", ctx.Err())
			//阻塞，等待锁到达了上限时间
		case <-timeoutCh:
			return fmt.Errorf("block waiting time out, err: %w", ErrLockAcquiredByOthers)
		default:
		}

		//尝试取锁
		err := r.tryLock(ctx)
		if err == nil {
			return nil
		}

		//为不可再尝试类型的错误返回
		if ok := IsRetryableErr(err); !ok {
			return err
		}
	}
	return nil
}

// 解锁
func (r *RedisLock) Unlock(ctx context.Context) error {
	// 开始事务
	err := r.client.Watch(ctx, func(tx *redis.Tx) error {
		// 监视锁的键
		value, err := tx.Get(ctx, r.getLockKey()).Result()
		if err != nil {
			// 如果键不存在或其他错误发生，直接返回错误
			if err == redis.Nil {
				return errors.New("lock does not exist")
			}
			return err
		}

		// 检查是否持有锁
		if value != r.token {
			return errors.New("can not unlock without ownership of lock")
		}

		// 开始事务：删除锁
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Del(ctx, r.getLockKey())
			return nil
		})

		return err
	}, r.getLockKey()) // Watch 锁的键

	if err != nil {
		return err
	}

	return nil
}
