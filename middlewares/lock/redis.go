package lock

import (
	"errors"
	"fmt"
	"github.com/bitstormhub/bitstorm/pkg/middlewares/cache"
	"github.com/google/uuid"
	"golang.org/x/net/context"
	"golang.org/x/sync/singleflight"
	"sync/atomic"
	"time"
)

const (
	// 默认的分布式锁过期时间
	defaultLockExpireSeconds = 3
	redisLockKeyPrefix       = "redisLock:"
)

var (
	ErrLockAcquiredByOthers = errors.New("lock is acquired by others")
	ErrDeleteLockFailure    = errors.New("unlock failed and lock is acquired by others")
	ErrDelayExpire          = errors.New("delayExpire failed")
)

// 基于 redis 实现的分布式锁，不可重入，但保证了对称性
type RedisLock struct {
	lockOptions
	key   string
	token string
	// watch Dog运作标识
	runningDog int32
	// 停止看门狗
	client  cache.Client
	stopDog context.CancelFunc
	g       singleflight.Group
}

type lockOptions struct {
	isBlock              bool
	watchDogWorkStepTime time.Duration // watch Dog工作时间间隙
	blockWaitingSecond   int64         // 阻塞等待加锁时间
	expireTimeSecond     int64         // 锁过期时间
	watchDogMode         bool          // 是否开启看门狗
}

type LockOption func(*lockOptions)

// WithLock 设置阻塞模式
func WithBlock() LockOption {
	return func(o *lockOptions) {
		o.isBlock = true
	}
}

// 设置阻塞最长等待时间
func WithBlockWaitingSeconds(waitingSeconds int64) LockOption {
	return func(o *lockOptions) {
		o.blockWaitingSecond = waitingSeconds
	}
}

// 设置过期时间
func WithExpireSeconds(expireTime int64) LockOption {
	return func(o *lockOptions) {
		o.expireTimeSecond = expireTime
	}
}

// 启动看门狗
func WithWatchDogMode() LockOption {
	return func(o *lockOptions) {
		o.watchDogMode = true
	}
}

// repairLock 修复分布式锁选项
func repairLock(o *lockOptions) {
	if o.isBlock && o.blockWaitingSecond <= 0 {
		// 默认阻塞等待时间上限为 5 秒
		o.blockWaitingSecond = 5
	}

	if o.expireTimeSecond <= 0 {
		// 关闭看门狗，如果没有设置过期时间
		o.watchDogMode = false
	}

	if o.watchDogMode {
		// 用户开启了看门狗
		// 将看门狗时间设置为 过期时间的 1/4
		o.watchDogWorkStepTime = time.Duration(o.expireTimeSecond) * time.Second / 3
	}

}

func isRetryableErr(err error) bool {
	return errors.Is(err, ErrLockAcquiredByOthers)
}

func NewRedisLock(key string, opts ...LockOption) *RedisLock {
	id, _ := uuid.NewUUID()
	r := RedisLock{
		key:   key,
		token: id.String(),
	}

	for _, opt := range opts {
		opt(&r.lockOptions)
	}

	repairLock(&r.lockOptions)
	return &r
}

// Lock 加锁
func (r *RedisLock) Lock(ctx context.Context) (err error) {
	defer func() {
		if err != nil {
			return
		}
		// 加锁成功的情况下，会启动看门狗
		// 关于该锁本身是不可重入的，所以不会出现同一把锁下看门狗重复启动的情况
		r.watchDog(ctx)
	}()

	// 不管是不是阻塞模式，都要先获取一次锁
	err = r.tryLock(ctx)
	if err == nil {
		// 加锁成功
		return nil
	}

	// 非阻塞模式加锁失败直接返回错误
	if !r.isBlock {
		return err
	}

	// 判断错误是否可以允许重试，不可允许的类型则直接返回错误
	if !isRetryableErr(err) {
		return err
	}

	// 基于阻塞模式持续轮询取锁
	err = r.blockingLock2(ctx)
	return
}

// Lock 加锁
func (r *RedisLock) SingleFlightLock(ctx context.Context) (err error) {
	defer func() {
		if err != nil {
			return
		}
		// 加锁成功的情况下，会启动看门狗
		// 关于该锁本身是不可重入的，所以不会出现同一把锁下看门狗重复启动的情况
		r.watchDog(ctx)
	}()

	// 不管是不是阻塞模式，都要先获取一次锁
	err = r.tryLock(ctx)
	if err == nil {
		// 加锁成功
		return nil
	}

	// 非阻塞模式加锁失败直接返回错误
	if !r.isBlock {
		return err
	}

	// 判断错误是否可以允许重试，不可允许的类型则直接返回错误
	if !isRetryableErr(err) {
		return err
	}

	// 基于阻塞模式持续轮询取锁
	err = r.blockingLock1(ctx)
	return
}

func (r *RedisLock) tryLock(ctx context.Context) error {
	// 首先查询锁是否属于自己
	result, err := cache.GetRedisCli().SetNX(ctx, r.key, r.token, time.Duration(r.expireTimeSecond)*time.Second)
	if err != nil {
		return err
	}

	// 为false，加锁失败，已经有锁
	if !result {
		return ErrLockAcquiredByOthers
	}

	return nil
}

// 启动看门狗
func (r *RedisLock) watchDog(ctx context.Context) {
	// 1. 非看门狗模式，不处理
	if !r.watchDogMode {
		return
	}

	// 2. 确保之前启动的看门狗已经正常回收
	for !atomic.CompareAndSwapInt32(&r.runningDog, 0, 1) {
	}

	// 3. 启动看门狗
	ctx, r.stopDog = context.WithCancel(ctx)
	go func() {
		defer func() {
			atomic.StoreInt32(&r.runningDog, 0)
		}()
		r.runWatchDog(ctx)
	}()
}

// runWatchDog 看门狗运作
func (r *RedisLock) runWatchDog(ctx context.Context) {
	ticker := time.NewTicker(r.watchDogWorkStepTime)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// 看门狗负责在用户未显式解锁时，持续为分布式锁进行续期
		// 通过 lua 脚本，延期之前会确保保证锁仍然属于自己
		_ = r.delayExpire(ctx, r.expireTimeSecond)
	}
}

// 更新锁的过期时间，基于 lua 脚本实现操作原子性
func (r *RedisLock) delayExpire(ctx context.Context, expireSeconds int64) error {
	result, err := cache.GetRedisCli().EvalBool(ctx, LuaCheckAndExpireDistributionLock, []string{r.key}, []interface{}{r.token, expireSeconds})
	if err != nil {
		return nil
	}

	if !result {
		return ErrDelayExpire
	}

	return nil
}

// singleflight模式
func (r *RedisLock) blockingLock1(ctx context.Context) error {
	// 阻塞模式等锁时间上限
	timeoutCh := time.After(time.Duration(r.blockWaitingSecond) * time.Second)
	// 轮询 ticker，每隔 50 ms 尝试取锁一次
	ticker := time.NewTicker(time.Duration(50) * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		select {
		// ctx 终止了
		case <-ctx.Done():
			return fmt.Errorf("lock failed, ctx timeout, err: %w", ctx.Err())
			// 阻塞等锁达到上限时间，返回
		case <-timeoutCh:
			return fmt.Errorf("block waiting time out, err: %w", ErrLockAcquiredByOthers)
		// 放行
		default:
		}

		err := func() error {
			flag := false
			// 尝试取锁
			ch := r.g.DoChan(r.key, func() (interface{}, error) {
				flag = true // 代表执行了
				err := r.tryLock(ctx)
				if err != nil {
					return nil, err
				}
				return nil, nil
			})

			select {
			case res := <-ch:
				if flag { // 当前goroutine拿的锁
					r.g.Forget(r.key)
					if res.Err != nil {
						return res.Err
					}
					return nil
				}
			case <-ctx.Done():
				ctx.Err()
			}

			return nil
		}()

		// 不可重试类型的错误，直接返回
		if !isRetryableErr(err) {
			return err
		}
	}

	// 不可达
	return nil
}

// 非singleflight模式
func (r *RedisLock) blockingLock2(ctx context.Context) error {
	// 阻塞模式等锁时间上限
	timeoutCh := time.After(time.Duration(r.blockWaitingSecond) * time.Second)
	// 轮询 ticker，每隔 50 ms 尝试取锁一次
	ticker := time.NewTicker(time.Duration(50) * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		select {
		// ctx 终止了
		case <-ctx.Done():
			return fmt.Errorf("lock failed, ctx timeout, err: %w", ctx.Err())
			// 阻塞等锁达到上限时间，返回
		case <-timeoutCh:
			return fmt.Errorf("block waiting time out, err: %w", ErrLockAcquiredByOthers)
		// 放行
		default:
		}

		err := r.tryLock(ctx)

		// 不可重试类型的错误，直接返回
		if !isRetryableErr(err) {
			return err
		}
	}

	// 不可达
	return nil
}

// Unlock 解锁. 基于 lua 脚本实现操作原子性.
func (r *RedisLock) Unlock(ctx context.Context) error {
	defer func() {
		// 停止看门狗
		if r.stopDog != nil {
			r.stopDog()
		}
	}()

	result, err := cache.GetRedisCli().EvalBool(ctx, LuaCheckAndDeleteDistributionLock, []string{r.key}, r.token)
	if err != nil {
		return err
	}

	// 释放成功则为true
	if !result {
		// 释放失败，不是自己的锁
		return ErrDeleteLockFailure
	}

	return nil
}

func (r *RedisLock) getLockKey() string {
	return redisLockKeyPrefix + r.key
}

// LuaCheckAndDeleteDistributionLock 判断是否拥有分布式锁的归属权，是则删除
const LuaCheckAndDeleteDistributionLock = `
  local lockerKey = KEYS[1]
  local targetToken = ARGV[1]
  local getToken = redis.call('get',lockerKey)
  if (not getToken or getToken ~= targetToken) then
    return 0
	else
		return redis.call('del',lockerKey)
  end
`

// LuaCheckAndExpireDistributionLock 判断是否拥有分布式锁的归属权，是则设置过期时间
const LuaCheckAndExpireDistributionLock = `
  local lockerKey = KEYS[1]
  local targetToken = ARGV[1]
  local duration = ARGV[2]
  local getToken = redis.call('get',lockerKey)
  if (not getToken or getToken ~= targetToken) then
    return 0
	else
		return redis.call('expire',lockerKey,duration)
  end
`
