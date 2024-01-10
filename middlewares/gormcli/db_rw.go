package gormcli

import (
	"fmt"
	"github.com/bitstormhub/bitstorm/pkg/middlewares/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
	"time"
)

// 读写分离配置
type RwOptions struct {
	masterAddr     string          // 主库地址，格式是 IP:PORT
	masterUser     string          // 主库用户名
	masterPassword string          // 主库密码
	masterDataBase string          // 主库db名
	slaveAddr      []string        // 从库地址，格式是 IP:PORT
	slaveUser      []string        // 从库用户名(顺序和地址一致)
	slavePassword  []string        // 从库密码(顺序和地址一致)
	slaveDataBase  []string        // 从库db名(顺序和地址一致)
	rwMaxIdleConn  int             // 最大空闲连接数
	rwMaxOpenConn  int             // 最大打开的连接数
	rwMaxIdleTime  int64           // 连接最大空闲时间
	rwLogger       *log.GormLogger // 日志输出
}

type RwOption func(*RwOptions)

func WithMasterAddr(addr string) RwOption {
	return func(o *RwOptions) {
		o.masterAddr = addr
	}
}

func WithMasterUser(user string) RwOption {
	return func(o *RwOptions) {
		o.masterUser = user
	}
}

func WithMasterPassword(password string) RwOption {
	return func(o *RwOptions) {
		o.masterPassword = password
	}
}

func WithMasterDataBase(dataBase string) RwOption {
	return func(o *RwOptions) {
		o.masterDataBase = dataBase
	}
}

func WithSlaveAddr(addr []string) RwOption {
	return func(o *RwOptions) {
		o.slaveAddr = addr
	}
}

func WithSlaveUser(user []string) RwOption {
	return func(o *RwOptions) {
		o.slaveUser = user
	}
}

func WithSlavePassword(password []string) RwOption {
	return func(o *RwOptions) {
		o.slavePassword = password
	}
}

func WithSlaveDataBase(dataBase []string) RwOption {
	return func(o *RwOptions) {
		o.slaveDataBase = dataBase
	}
}

func WithRwMaxIdleConn(maxIdleConn int) RwOption {
	return func(o *RwOptions) {
		o.rwMaxIdleConn = maxIdleConn
	}
}

func WithRwMaxOpenConn(maxOpenConn int) RwOption {
	return func(o *RwOptions) {
		o.rwMaxOpenConn = maxOpenConn
	}
}

// 设置最大空闲连接时间，设置的数字，内部会转换成秒数
func WithRwMaxIdleTime(maxIdleTime int64) RwOption {
	return func(o *RwOptions) {
		o.rwMaxIdleTime = maxIdleTime
	}
}

func WithRwLogger(logger *log.GormLogger) RwOption {
	return func(o *RwOptions) {
		o.rwLogger = logger
	}
}

func InitMasterAndSlave(opts ...RwOption) {
	db = newRwDB(newRwOptions(opts...))
}

func newRwOptions(opts ...RwOption) RwOptions {
	options := RwOptions{
		masterAddr:     "127.0.0.1:3306",
		masterUser:     "root",
		masterPassword: "root",
		masterDataBase: "bitstorm",
		slaveAddr:      []string{"127.0.0.1:3307", "127.0.0.1:3308"},
		slaveUser:      []string{"root", "root"},
		slavePassword:  []string{"root", "root"},
		slaveDataBase:  []string{"bitstorm", "bitstorm"},
		rwMaxIdleConn:  10,
		rwMaxOpenConn:  10,
		rwMaxIdleTime:  10,
	}

	for _, opt := range opts {
		opt(&options)
	}

	return options
}
func newRwDB(options RwOptions) *gorm.DB {
	var slaveDsn []string
	masterDsn := fmt.Sprintf(dsn, options.masterUser,
		options.masterPassword, options.masterAddr, options.masterDataBase)
	for i := range options.slaveAddr {
		slaveDsn = append(slaveDsn, fmt.Sprintf(dsn, options.slaveUser[i],
			options.slavePassword[i], options.slaveAddr[i], options.slaveDataBase[i]))
	}

	db, err := gorm.Open(mysql.Open(masterDsn), &gorm.Config{
		Logger: options.rwLogger,
	})

	if err != nil {
		panic("failed to connect master database")
	}

	var replicas []gorm.Dialector

	for _, slave := range slaveDsn {
		cfg := mysql.Config{
			DSN: slave,
		}
		replicas = append(replicas, mysql.New(cfg))
	}

	err = db.Use(dbresolver.Register(dbresolver.Config{
		Sources: []gorm.Dialector{mysql.New(mysql.Config{
			DSN: masterDsn,
		})},
		Replicas: replicas,
		Policy:   dbresolver.RandomPolicy{},
	}).
		SetMaxIdleConns(options.rwMaxIdleConn).
		SetConnMaxLifetime(time.Duration(options.rwMaxIdleTime * int64(time.Second))).
		SetMaxOpenConns(options.rwMaxOpenConn),
	)

	if err != nil {
		panic(err)
	}

	return db
}
