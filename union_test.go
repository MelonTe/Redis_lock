package redis_lock

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func doSomething(t *testing.T, ID int, rl *RedisLock, ctx context.Context) {
	t.Logf("goroutineID:%v tring to get the key...", ID)
	err := rl.Lock(ctx)
	if err != nil {
		t.Log(err)
		return
	}
	t.Logf("goroutineID:%v get key success", ID)
	t.Logf("goroutineID:%v need 2s to finish the mission", ID)
	time.Sleep(2 * time.Second)
	t.Logf("goroutineID:%v mission success, try to unlock...", ID)
	err = rl.Unlock(ctx)
	if err != nil {
		t.Logf("unlock failed, err:%v", err)
		return
	}
	t.Logf("goroutineID:%v unlock success, finish.", ID)
}
func TestLock(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		DB:       0,
		Addr:     "127.0.0.1:6379",
		Password: "MelonTe",
	})
	ctx := context.Background()
	rl := NewRedisLock("test_key", rdb, WithBlock(), WithWaitingSeconds(20))
	for i := 1; i <= 3; i++ {
		go doSomething(t, i, rl, ctx)
	}
	time.Sleep(100 * time.Second)
	return
}
