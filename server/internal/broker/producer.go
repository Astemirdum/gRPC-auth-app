package broker

import (
	"encoding/json"
	"time"

	"github.com/Astemirdum/user-app/server/internal/config"
	"github.com/Astemirdum/user-app/server/models"
	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
)

type Producer struct {
	prod  sarama.AsyncProducer
	topic string
}

func NewProducer(kcfg *config.Kafka) (*Producer, error) {
	cfg := sarama.NewConfig()
	cfg.Producer.Partitioner = sarama.NewRandomPartitioner
	cfg.Producer.Return.Successes = true
	brokers := []string{kcfg.Addr}
	p, err := sarama.NewAsyncProducer(brokers, cfg)
	if err != nil {
		return nil, err
	}
	prod := &Producer{
		prod:  p,
		topic: kcfg.Topic,
	}

	go func() {
		for err := range p.Errors() {
			logrus.Debugf("Msg err: %v", err)
		}
	}()

	go func() {
		for suc := range p.Successes() {
			logrus.Debugf("Msg written Patrition: %d. Ossfet: %d",
				suc.Partition, suc.Offset)
		}
	}()

	return prod, nil
}

func (p *Producer) Publish(user *models.User) error {
	userLog := models.UserLog{
		Id:        int32(user.Id),
		Email:     user.Email,
		Password:  user.Password,
		Timestamp: time.Now().UTC().Unix(),
	}
	usLog, err := json.Marshal(userLog)
	if err != nil {
		return err
	}
	p.sendMessage(p.topic, usLog)
	return nil
}

func (p *Producer) sendMessage(topic string, message []byte) {
	msg := &sarama.ProducerMessage{
		Topic:     topic,
		Partition: -1,
		Value:     sarama.ByteEncoder(message),
	}
	p.prod.Input() <- msg
}

func (p *Producer) Close() error {
	return p.prod.Close()
}
