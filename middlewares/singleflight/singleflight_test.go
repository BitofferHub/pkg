package singleflight

import (
	"golang.org/x/sync/singleflight"
	"log"
	"sync"
	"testing"
)

var (
	sfKey1 = "key1"
	wg     *sync.WaitGroup
	sf     singleflight.Group
	nums   = 1000000

	mutex sync.Mutex
)

func TestSingleFlight(t *testing.T) {

	getValueService()
}

func getValueService() { // service
	var val int
	wg = &sync.WaitGroup{}
	wg.Add(nums)
	for idx := 0; idx < nums; idx++ { // 模拟多协程同时请求
		go func(idx int) {
			defer wg.Done()
			getValueBySingleFlight(idx, &val) // 简化代码，不处理error
		}(idx)
	}
	wg.Wait()
	log.Println("val: ", val)
	return
}

// getValueBySingleFlight 使用singleFlight取cacheKey对应的value值
func getValueBySingleFlight(idx int, val *int) {
	var loop = 0
	for {
		flag := false
		// 调用singleFlight的Do()方法
		ch := sf.DoChan(sfKey1, func() (ret interface{}, err error) {
			flag = true

			mutex.Lock()
			*val++
			mutex.Unlock()
			return nil, nil
		})

		select {
		case <-ch:
			if flag {
				// 是当前goroutine
				sf.Forget(sfKey1)
				return
			}
		}

		loop++
	}
}
