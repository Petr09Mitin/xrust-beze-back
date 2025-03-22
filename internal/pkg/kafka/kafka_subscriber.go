package infrakafka

import (
	"github.com/IBM/sarama"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"time"
)

const (
	kafkaVersion = "7.6.0"
	kafkaAdress  = "kafka_xb:9092"
)

func NewKafkaSubscriber() (message.Subscriber, error) {
	saramaConfig := sarama.NewConfig()
	saramaVersion, err := sarama.ParseKafkaVersion(kafkaVersion)
	if err != nil {
		return nil, err
	}
	saramaConfig.Version = saramaVersion
	saramaConfig.Consumer.Fetch.Default = 1024 * 1024
	saramaConfig.Consumer.Offsets.AutoCommit.Enable = true
	saramaConfig.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second

	kafkaSubscriber, err := kafka.NewSubscriber(
		kafka.SubscriberConfig{
			Brokers:       []string{kafkaAdress},
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
