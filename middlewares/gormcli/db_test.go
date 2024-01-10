package gormcli

import (
	"context"
	"github.com/BitofferHub/pkg/constant"
	"github.com/BitofferHub/pkg/middlewares/log"
	"testing"
	"time"
)

type UrlMap struct {
	ID        int64 `gorm:"primaryKey"`
	LongUrl   string
	ShortUrl  string
	CreatedAt time.Time
}

func (UrlMap) TableName() string {
	return "url_map"
}

func TestDB(t *testing.T) {
	// 初始化日志
	log.Init(log.WithLogLevel("debug"),
		log.WithFileName("bitstorm.log"),
		log.WithMaxSize(100),
		log.WithMaxBackups(3),
		log.WithLogPath("./log"),
		log.WithConsole(true),
	)

	Init(WithAddr("127.0.0.1:3306"),
		WithUser("root"),
		WithPassword("root"),
		WithDataBase("shorturlx"),
		WithMaxIdleConn(20),
		WithMaxOpenConn(100),
		// 100s
		WithMaxIdleTime(100),
		// 设置慢查询阈值(设置10，就是10ms)
		WithSlowThresholdMillisecond(10),
	)

	db := GetDB()
	if db == nil {
		t.Error("db is nil")
	}
	urlMap := &UrlMap{
		LongUrl:  "hts://wwfdsfom",
		ShortUrl: "123ba",
	}
	err := db.WithContext(context.WithValue(context.Background(), constant.TraceID, "bitstorm success")).Create(urlMap).Error
	if err != nil {
		t.Error(err)
	}
	log.Errorf("bitstorm test %v", "success")

	Close()
}

func TestRwDB(t *testing.T) {
	// 初始化日志
	log.Init(log.WithLogLevel("debug"),
		log.WithFileName("bitstorm.log"),
		log.WithMaxSize(100),
		log.WithMaxBackups(3),
		log.WithLogPath("./log"),
		log.WithConsole(true),
	)

	InitMasterAndSlave(WithMasterAddr("192.168.5.51:3306"),
		WithMasterUser("root"),
		WithMasterPassword("123456"),
		WithMasterDataBase("shorturlx"),
		WithSlaveAddr([]string{"192.168.5.52:3307", "192.168.5.52:3308", "192.168.5.52:3309"}),
		WithSlavePassword([]string{"123456", "123456", "123456"}),
		WithSlaveUser([]string{"root", "root", "root"}),
		WithSlaveDataBase([]string{"shorturlx", "shorturlx", "shorturlx"}),
		WithRwMaxIdleConn(10),
		WithRwMaxOpenConn(10),
		WithRwMaxOpenConn(100),
		// 100s
		WithRwMaxIdleTime(100),
		// 带上日志
		WithRwLogger(log.NewGormLogger(10)),
	)

	var err error
	db := GetDB()
	if db == nil {
		t.Error("db is nil")
	}
	urlMap := &UrlMap{
		LongUrl:  "hts://wwfdsfom",
		ShortUrl: "123ba",
	}
	err = db.WithContext(context.WithValue(context.Background(), constant.TraceID, "bitstorm success")).Create(urlMap).Error
	if err != nil {
		t.Error(err)
	}

	// 查询
	var urlMap2 UrlMap
	err = db.WithContext(context.WithValue(context.Background(), constant.TraceID, "bitstorm success")).Where("short_url = ?", "123ba").First(&urlMap2).Error
	if err != nil {
		t.Error(err)
	}

	t.Log(urlMap2)

	Close()
}

func TestShardingSphere(t *testing.T) {

}
