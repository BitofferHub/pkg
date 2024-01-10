package cache

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"math/rand"
	"testing"
	"time"
)

func InitTest() {

	Init(WithAddr("127.0.0.1:6379"),
		WithDB(0),
		WithPoolSize(10),
		WithPassWord(""),
	)
}

func TestCache(t *testing.T) {
	InitTest()
	// 生成一个随机数
	randNum := rand.Intn(100)
	k := fmt.Sprintf("key_%d", randNum)
	v := fmt.Sprintf("value_%d", randNum)
	// 设置缓存
	err := GetRedisCli().Set(context.Background(), k, v, 1*time.Second)
	if err != nil {
		panic(err)
	}
	// 获取缓存
	val, exist, err := GetRedisCli().Get(context.Background(), k)
	if !exist {
		panic(err)
	}

	if err != nil {
		panic(err)
	}

	if val != v {
		panic("value not equal")
	}

	// 关闭redis 连接
	GetRedisCli().Close()
}

func TestPipeline(t *testing.T) {
	InitTest()
	// 生成一个随机数
	randNum := rand.Intn(100)
	k := fmt.Sprintf("key_%d", randNum)

	var res1 *redis.StatusCmd
	var res3 *redis.StringCmd
	var res2 *redis.IntCmd
	err := GetRedisCli().Pipeline(context.Background(), func(pipe redis.Pipeliner) error {
		res1 = pipe.Set(context.Background(), k, 100, 10*time.Second)
		res2 = pipe.Incr(context.Background(), k)
		res3 = pipe.Get(context.Background(), k)
		return nil
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(res1.Val(), res2.Val(), res3.Val())
}

func TestBloomFilter(t *testing.T) {
	InitTest()

	// 占13G内存
	// var size int64 = 100000000 * 100 // 100亿

	// GetRedisCli().BFReserve(context.Background(), "test2", 0.01, size)
	//
	// 放元素
	for i := 0; i < 3; i++ {
		res, err := GetRedisCli().BFAdd(context.Background(), "test2", fmt.Sprintf("test%d", i))
		t.Log(res, err)
	}

}

func TestLua(t *testing.T) {
	// 注意事项:
	// 1. 在lua中，多返回值，使用{}来返回
	// 2. 不要返回true和false，直接用其他数字来代替
	// 3. lua返回浮点数，会舍弃小数，最终还是int64

	InitTest()
	// 准备执行的 Lua 脚本
	script := `
		return {"value1", 123, "value3", 11.11, true}
	`

	res, err := GetRedisCli().EvalResults(context.Background(), script, nil)
	if err != nil {
		panic(err)
	}
	// 解析返回值
	fmt.Println(res)
	values := res.([]interface{})

	// 获取每个返回值
	// 更新详细信息看这里
	// https://redis.uptrace.dev/zh/guide/lua-scripting.html#lua-%E5%92%8C-go-%E7%B1%BB%E5%9E%8B
	result1 := values[0].(string)
	result2 := values[1].(int64)
	result3 := values[2].(string)
	result4 := values[3].(int64) // lua不返回小数，舍弃小数，float转换成int64
	result5 := values[4].(int64) // lua 返回的true，会转换成int64(1)

	fmt.Println(result1, result2, result3, result4, result5)
}
