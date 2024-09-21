package redis_lock

type LockOptions struct {
	//是否为阻塞式分布式锁
	isBlock             bool
	blockWaitingSeconds int64
	expireSeconds       int64
}
type LockOption func(*LockOptions)

//该方法用于将锁设置为阻塞锁
func WithBlock() LockOption {
	return func(o *LockOptions) {
		o.isBlock = true
	}
}

//该方法用于设置阻塞锁的等待时间
func WithWaitingSeconds(waitingSeconds int64) LockOption {
	return func(o *LockOptions) {
		o.blockWaitingSeconds = waitingSeconds
	}
}

//该方法用于设置锁的过期时间
func WithExpireSeconds(expireSeconds int64) LockOption {
	return func(o *LockOptions) {
		o.expireSeconds = expireSeconds
	}
}

func repairLock(o *LockOptions) {
	if o.isBlock && o.blockWaitingSeconds <= 0 {
		//默认阻塞时间为5s
		o.blockWaitingSeconds = 5
	}

	//分布式锁默认超时时间为30s
	if o.expireSeconds <= 0 {
		o.expireSeconds = 30
	}
}
