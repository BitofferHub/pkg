package gormcli

import (
	"context"
	"fmt"
	"github.com/BitofferHub/pkg/middlewares/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"reflect"
	"time"
)

type ctxTransactionKey struct {
}

var (
	db  *gorm.DB
	dsn = "%s:%s@(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local"
)

type Options struct {
	addr                     string // 地址，格式是 IP:PORT
	user                     string // 用户名
	password                 string // 密码
	dataBase                 string // db名
	maxIdleConn              int    // 最大空闲连接数
	maxOpenConn              int    // 最大打开的连接数
	maxIdleTime              int64  // 连接最大空闲时间
	slowThresholdMillisecond int64  // 慢查询SQL的阈值
}

type Option func(*Options)

func WithSlowThresholdMillisecond(slowThresholdMillisecond int64) Option {
	return func(o *Options) {
		o.slowThresholdMillisecond = slowThresholdMillisecond
	}
}

func WithAddr(addr string) Option {
	return func(o *Options) {
		o.addr = addr
	}
}

func WithUser(user string) Option {
	return func(o *Options) {
		o.user = user
	}
}

func WithPassword(password string) Option {
	return func(o *Options) {
		o.password = password
	}
}

func WithDataBase(dataBase string) Option {
	return func(o *Options) {
		o.dataBase = dataBase
	}
}

func WithMaxIdleConn(maxIdleConn int) Option {
	return func(o *Options) {
		o.maxIdleConn = maxIdleConn
	}
}

func WithMaxOpenConn(maxOpenConn int) Option {
	return func(o *Options) {
		o.maxOpenConn = maxOpenConn
	}
}

// 设置最大空闲连接时间，设置的数字，内部会转换成秒数
func WithMaxIdleTime(maxIdleTime int64) Option {
	return func(o *Options) {
		o.maxIdleTime = maxIdleTime
	}
}

func newOptions(opts ...Option) Options {
	options := Options{
		addr:                     "127.0.0.1:3306",
		user:                     "root",
		password:                 "root",
		dataBase:                 "bitstorm",
		maxIdleConn:              10,
		maxOpenConn:              10,
		maxIdleTime:              10,
		slowThresholdMillisecond: 0,
	}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

func Init(options ...Option) {
	newDB(newOptions(options...))
}

func newDB(options Options) {
	var err error
	connArgs := fmt.Sprintf(dsn, options.user,
		options.password, options.addr, options.dataBase)
	if options.slowThresholdMillisecond != 0 {
		gormLogger := log.NewGormLogger(options.slowThresholdMillisecond)
		db, err = gorm.Open(mysql.Open(connArgs), &gorm.Config{
			Logger: gormLogger,
		})
	} else {
		// 没有配置慢查询阈值，直接不输出日志
		// 不输出到终端
		db, err = gorm.Open(mysql.Open(connArgs), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
	}

	if err != nil {
		panic("failed to connect database")
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic("fetch gormCli connection err:" + err.Error())
	}

	sqlDB.SetMaxIdleConns(options.maxIdleConn)                                        // 设置最大空闲连接
	sqlDB.SetMaxOpenConns(options.maxOpenConn)                                        // 设置最大打开的连接
	sqlDB.SetConnMaxLifetime(time.Duration(options.maxIdleTime * int64(time.Second))) // 设置空闲时间为(s)
}

// GetDB 获取数据库连接
func GetDB() *gorm.DB {
	return db
}

func Close() {
	if db != nil {
		sqlDB, err := db.DB()
		if err != nil {
			log.Errorf("close gormcli err:%v", err)
		}
		sqlDB.Close()
	}
}

func Transaction(ctx context.Context, fc func(txctx context.Context) error) error {
	db := GetDB().WithContext(ctx)

	return db.Transaction(func(tx *gorm.DB) error {
		txctx := CtxWithTransaction(ctx, tx)
		return fc(txctx)
	})
}

func GetDBFromCtx(ctx context.Context) *gorm.DB {
	iface := ctx.Value(ctxTransactionKey{})

	if iface != nil {
		tx, ok := iface.(*gorm.DB)
		if !ok {
			log.ErrorContextf(ctx, "unexpect context value type: %s", reflect.TypeOf(tx))
			return nil
		}

		return tx
	}

	return GetDB().WithContext(ctx)
}

func CtxWithTransaction(ctx context.Context, tx *gorm.DB) context.Context {
	c := context.WithValue(ctx, ctxTransactionKey{}, tx)
	return c
}
