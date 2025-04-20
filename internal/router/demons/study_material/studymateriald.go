package study_materiald

import (
	"context"
	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/config"
	infrakafka "github.com/Petr09Mitin/xrust-beze-back/internal/pkg/kafka"
	"github.com/Petr09Mitin/xrust-beze-back/internal/services/study_material"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/rs/zerolog"
)

type StudyMaterialD struct {
	studyMaterialService study_material.StudyMaterialService
	subTopicID           string
	router               *message.Router
	sub                  message.Subscriber
	logger               zerolog.Logger
}

func NewStudyMaterialD(
	studyMaterialService study_material.StudyMaterialService,
	subTopicID string,
	kafkaCfg *config.Kafka,
	logger zerolog.Logger,
) (*StudyMaterialD, error) {
	router, err := infrakafka.NewBrokerRouter()
	if err != nil {
		return nil, err
	}
	sub, err := infrakafka.NewKafkaSubscriber(kafkaCfg)
	if err != nil {
		return nil, err
	}
	return &StudyMaterialD{
		studyMaterialService: studyMaterialService,
		subTopicID:           subTopicID,
		router:               router,
		sub:                  sub,
		logger:               logger,
	}, nil
}

func (s *StudyMaterialD) Run(ctx context.Context) error {
	s.registerHandler()
	return s.router.Run(ctx)
}

func (s *StudyMaterialD) GracefulStop() error {
	return s.router.Close()
}

func (s *StudyMaterialD) registerHandler() {
	s.router.AddNoPublisherHandler(
		"study_material_handler",
		s.subTopicID,
		s.sub,
		s.handleMessage,
	)
}

func (s *StudyMaterialD) handleMessage(msg *message.Message) error {
	return nil
}
