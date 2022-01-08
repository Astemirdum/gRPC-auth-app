package server

import (
	"context"
	"strconv"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	Prod *kafka.Writer
}

func NewKafkaProducer(topic, brokerAddress string) *Producer {
	return &Producer{kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{brokerAddress},
		Topic:   topic,
	}),
	}
}

func (p *Producer) Produce(ctx context.Context, uid int) error {
	err := p.Prod.WriteMessages(ctx, kafka.Message{
		Key:   []byte(strconv.Itoa(uid)),
		Value: []byte("user log #" + strconv.Itoa(uid)),
	})
	return err
}
