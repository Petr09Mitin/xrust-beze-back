package infrakafka

import (
	"errors"
	"github.com/IBM/sarama"
	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/config"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"time"
)

func NewKafkaSubscriber(cfg *config.Kafka) (message.Subscriber, error) {
	if cfg == nil {
		return nil, errors.New("kafka subscriber config is nil")
	}
	saramaConfig := sarama.NewConfig()
	saramaVersion, err := sarama.ParseKafkaVersion(cfg.Version)
	if err != nil {
		return nil, err
	}
	saramaConfig.Version = saramaVersion
	saramaConfig.Consumer.Fetch.Default = 1024 * 1024
	saramaConfig.Consumer.Offsets.AutoCommit.Enable = true
	saramaConfig.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second

	kafkaSubscriber, err := kafka.NewSubscriber(
		kafka.SubscriberConfig{
			Brokers:       cfg.Addresses,
			Unmarshaler:   kafka.DefaultMarshaler{},
			ConsumerGroup: watermill.NewUUID(),
			InitializeTopicDetails: &sarama.TopicDetail{
				NumPartitions:     1,
				ReplicationFactor: 2,
			},
			OverwriteSaramaConfig: saramaConfig,
		},
		logger,
	)
	if err != nil {
		return nil, err
	}

	return kafkaSubscriber, nil
}
