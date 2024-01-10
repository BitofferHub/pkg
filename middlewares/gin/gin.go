package gin

import (
	"github.com/gin-gonic/gin"
	"io"
)

func NewEngine(opts ...Option) *gin.Engine {
	opt := newOptions(opts...)

	if !opt.logger {
		gin.DefaultWriter = io.Discard
	}
	engine := gin.New()
	return engine
}

func newOptions(opts ...Option) Options {
	// 默认配置
	options := Options{
		logger: true,
	}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

type Options struct {
	logger bool
}

type Option func(*Options)

func WithLogger(logger bool) Option {
	return func(o *Options) {
		o.logger = logger
	}
}
