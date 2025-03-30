package infrakafka

import (
	"errors"
	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/config"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
)

var (
	logger = watermill.NewStdLogger(
		false,
		false,
	)
)

func NewKafkaPublisher(cfg *config.Kafka) (message.Publisher, error) {
	if cfg == nil {
		return nil, errors.New("kafka publisher config is nil")
	}
	kafkaPublisher, err := kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers:   cfg.Addresses,
			Marshaler: kafka.DefaultMarshaler{},
		},
		logger,
	)
	if err != nil {
		return nil, err
	}

	return kafkaPublisher, nil
}

func NewBrokerRouter() (*message.Router, error) {
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, err
	}
	router.AddMiddleware(
		middleware.CorrelationID,
		middleware.Timeout(time.Second*15),
		middleware.Recoverer,
	)
	return router, nil
}
