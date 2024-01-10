package lock

import (
	"context"
	"errors"
	"fmt"
	"github.com/BitofferHub/pkg/middlewares/cache"
	"github.com/smartystreets/goconvey/convey"
	"sync"
	"testing"
	"time"
)

func Init() {
	// 先初始化
	cache.Init(cache.WithAddr("127.0.0.1:6379"),
		cache.WithDB(0),
		cache.WithPoolSize(10),
		cache.WithPassWord(""),
	)
}

func TestLock(t *testing.T) {
	Init()

	convey.Convey("TestLock", t, func() {
		fmt.Println(5 * time.Second / 4)
		ctx := context.Background()
		// 过期时间为1秒，并开启续期
		lock1 := NewRedisLock("test_key", WithExpireSeconds(5), WithWatchDogMode())
		lock2 := NewRedisLock("test_key", WithExpireSeconds(1), WithBlock(), WithBlockWaitingSeconds(10))
		fmt.Println(lock1)
		fmt.Println(lock2)
		if err := lock1.Lock(ctx); err != nil {
			t.Error(err)
			return
		}

		go func() {
			time.Sleep(6 * time.Second)
			if err := lock1.Unlock(ctx); err != nil {
				t.Error()
			}
		}()

		if err := lock2.Lock(ctx); err != nil {
			t.Error(err)
		}

		time.Sleep(5 * time.Second)
	})
}

func TestBlockingLock(t *testing.T) {
	Init()

	convey.Convey("TestBlockingLock", t, func() {

		// 过期时间为1秒
		lock1 := NewRedisLock("test_key", WithExpireSeconds(1))
		// 阻塞等待时间为2秒
		lock2 := NewRedisLock("test_key", WithBlock(), WithBlockWaitingSeconds(2))

		ctx := context.Background()
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := lock1.Lock(ctx); err != nil {
				t.Error(err)
				return
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := lock2.Lock(ctx); err != nil {
				t.Error(err)
				return
			}
		}()

		wg.Wait()

		t.Log("success")
	})
}

func Test_nonblockingLock(t *testing.T) {
	Init()

	lock1 := NewRedisLock("test_key", WithExpireSeconds(1))
	lock2 := NewRedisLock("test_key")

	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := lock1.Lock(ctx); err != nil {
			t.Error(err)
			return
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := lock2.Lock(ctx); err == nil || !errors.Is(err, ErrLockAcquiredByOthers) {
			t.Errorf("got err: %v, expect: %v", err, ErrLockAcquiredByOthers)
			return
		}
	}()

	wg.Wait()
	t.Log("success")
}

func BenchmarkRedisLock(t *testing.B) {
	Init()

	var num = 100
	for i := 0; i < t.N; i++ {
		var count int
		var wg sync.WaitGroup
		lock := NewRedisLock("lock", WithExpireSeconds(2), WithBlock(), WithBlockWaitingSeconds(10))
		wg.Add(num)
		for j := 0; j < num; j++ {
			go func() {
				defer wg.Done()
				err := lock.Lock(context.Background())
				if err != nil {
					t.Error(err)
				}
				for i := 0; i < 100; i++ {
					count++
				}

				lock.Unlock(context.Background())
			}()
		}
		wg.Wait()
		t.Log(count)
	}

}

func BenchmarkSingleFlightLock(t *testing.B) {
	Init()

	var num = 1000
	for i := 0; i < t.N; i++ {
		var count int
		var wg sync.WaitGroup
		wg.Add(num)
		for j := 0; j < num; j++ {
			go func() {
				lock := NewRedisLock("lock", WithExpireSeconds(2), WithBlock(), WithBlockWaitingSeconds(10))
				defer wg.Done()
				err := lock.SingleFlightLock(context.Background())
				if err != nil {
					t.Error(err)
				}
				for i := 0; i < 100; i++ {
					count++
				}
				lock.Unlock(context.Background())
			}()
		}
		wg.Wait()
		t.Log(count)
	}
}

// BenchmarkRedisLock
// BenchmarkRedisLock-8   	       1	1971715600 ns/op
// PASS

// BenchmarkSingleFlightLock
// BenchmarkSingleFlightLock-8   	      18	  69731433 ns/op
// PASS
