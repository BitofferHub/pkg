package mq

import (
	"context"
	"github.com/BitofferHub/pkg/middlewares/log"
	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

func (kp *KafkaProducer) SendMessage(message []byte) error {
	return kp.writer.WriteMessages(context.Background(), kafka.Message{
		Value: message,
	})
}

func (kp *KafkaProducer) Close() {
	if err := kp.writer.Close(); err != nil {
		log.Errorf("Error closing producer: %v", err)
	}
}

func NewKafkaProducer(options ...Option) Producer {
	opts := newOptions(options...)

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(opts.brokers...),
		Topic:                  opts.topic,
		Balancer:               &kafka.LeastBytes{},          // 指定分区的balancer模式为最小字节分布
		RequiredAcks:           kafka.RequiredAcks(opts.ack), // ack模式
		Async:                  opts.async,                   // 是否开启异步模式（开启异步模式后，sendMessage 方法不会阻塞）
		AllowAutoTopicCreation: true,                         // 自动创建topic
	}

	return &KafkaProducer{writer: writer}
}

func NewKafkaConsumer(options ...Option) Consumer {
	opts := newOptions(options...)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   opts.brokers,
		GroupID:   opts.groupID, // 使用消费者组时，ReadMessage 会自动提交偏移量。
		Topic:     opts.topic,
		Partition: opts.partition,
	})
	// 设置消费偏移量
	reader.SetOffset(opts.offset)

	return &KafkaConsumer{reader: reader}
}

type KafkaConsumer struct {
	reader *kafka.Reader
}

func (kc *KafkaConsumer) ConsumeMessages(handler func([]byte) error) {
	for {
		message, err := kc.reader.FetchMessage(context.Background())
		if err != nil {
			log.Errorf("Error fetching message: %v", err)
			continue
		}

		err = handler(message.Value)
		if err != nil {
			log.Errorf("Error handling message: %v", err)
		}

		kc.reader.CommitMessages(context.Background(), message)
	}
}

func (kc *KafkaConsumer) Close() {
	if err := kc.reader.Close(); err != nil {
		log.Errorf("Error closing consumer: %v", err)
	}
}
