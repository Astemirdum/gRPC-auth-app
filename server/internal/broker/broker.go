package broker

import (
	"github.com/Astemirdum/user-app/server/internal/config"
	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
	"time"
)

type Broker struct {
	broker *sarama.Broker
}

func NewBroker(kcfg *config.Kafka) (*Broker, error) {
	broker := sarama.NewBroker(kcfg.Addr)
	cfg := sarama.NewConfig()

	if err := broker.Open(cfg); err != nil {
		return nil, err
	}
	br := &Broker{broker: broker}

	return br, br.createTopic(kcfg)
}

func (b *Broker) createTopic(kcfg *config.Kafka) error {

	topicDetail := &sarama.TopicDetail{}
	topicDetail.NumPartitions = int32(kcfg.PartitionNum)
	topicDetail.ReplicationFactor = 1
	topicDetail.ConfigEntries = make(map[string]*string)

	topicDetails := make(map[string]*sarama.TopicDetail)
	topicDetails[kcfg.Topic] = topicDetail
	request := sarama.CreateTopicsRequest{
		Timeout:      time.Second * 15,
		TopicDetails: topicDetails,
	}

	response, err := b.broker.CreateTopics(&request)
	if err != nil {
		logrus.Errorf("CreateTopics %v", err)
		return err
	}
	logrus.Debug("createTopic: response", response.TopicErrors)
	return nil
}

func (b *Broker) Close() error {
	return b.broker.Close()
}
