package mq

type Producer interface {
	SendMessage(message []byte) error
	Close()
}

type Consumer interface {
	ConsumeMessages(handler func([]byte) error)
	Close()
}

type Options struct {
	brokers []string // broker 地址
	async   bool     // 是否开启异步模式（开启异步模式后，sendMessage 方法不会阻塞）

	// 为0: 不需要等待broker的确认，提供了最低的延迟。但是最弱的持久性
	// 为1：需要等待leader的确认，较好的持久性，较低的延迟性。
	// 为2：需要等待follower的确认，持久性最好，延时性最差。
	// 三种机制随着数字递增，性能递减，可靠性递增。
	ack   int8
	topic string // topic

	groupID   string // consumer group id
	partition int    // partition

	offset int64 // 消费offset
}

type Option func(*Options)

func WithBrokers(brokers []string) Option {
	return func(o *Options) {
		o.brokers = brokers
	}
}

func WithAsync() Option {
	return func(o *Options) {
		o.async = true
	}
}

func WithAck(ack int8) Option {
	return func(o *Options) {
		o.ack = ack
	}
}

func WithTopic(topic string) Option {
	return func(o *Options) {
		o.topic = topic
	}
}

func WithOffset(offset int64) Option {
	return func(o *Options) {
		o.offset = offset
	}
}

func newOptions(opts ...Option) Options {
	options := Options{
		brokers: []string{"127.0.0.1:9092"},
	}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}
