package mq

import (
	"errors"
	"github.com/bitstormhub/bitstorm/pkg/middlewares/log"
	"testing"
)

func TestKafka(t *testing.T) {

}

func TestNewKafkaProducer(t *testing.T) {
	Init()

	producer := NewKafkaProducer(WithBrokers([]string{"127.0.0.1:29092"}), WithTopic("test"), WithAck(0), WithAsync())

	defer producer.Close()
	if err := producer.SendMessage([]byte("hello, world")); err != nil {
		t.Error(err)
	}
	if err := producer.SendMessage([]byte("hello, world")); err != nil {
		t.Error(err)
	}

}

func TestNewKafkaConsumer(t *testing.T) {
	Init()

	consumer := NewKafkaConsumer(WithBrokers([]string{"127.0.0.1:29092"}), WithTopic("test"), WithOffset(0))
	if consumer == nil {
		t.Fatal("nil consumer")
	}
	defer consumer.Close()

	consumer.ConsumeMessages(func(message []byte) error {
		t.Log("message is: ", string(message))
		if string(message) == "hello, world, end" {
			return errors.New("end")
		}
		return nil
	})

}

func Init() {
	// 初始化日志
	// 初始化日志
	log.Init(log.WithLogLevel("debug"),
		log.WithFileName("bitstorm.log"),
		log.WithMaxSize(100),
		log.WithMaxBackups(3),
		log.WithLogPath("./log"),
	)

}
