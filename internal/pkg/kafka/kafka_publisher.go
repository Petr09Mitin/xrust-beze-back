package infrakafka

import (
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

func NewKafkaPublisher() (message.Publisher, error) {
	kafkaPublisher, err := kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers:   []string{kafkaAdress},
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
