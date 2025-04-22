package study_materiald

import (
	"context"

	study_material_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/study_material"
	study_material_service "github.com/Petr09Mitin/xrust-beze-back/internal/services/study_material"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/rs/zerolog"
)

type StudyMaterialD struct {
	studyMaterialService study_material_service.StudyMaterialService
	subTopicID           string
	router               *message.Router
	sub                  message.Subscriber
	logger               zerolog.Logger
}

func NewStudyMaterialD(
	studyMaterialService study_material_service.StudyMaterialService,
	subTopicID string,
	router *message.Router,
	sub message.Subscriber,
	logger zerolog.Logger,
) *StudyMaterialD {
	return &StudyMaterialD{
		studyMaterialService: studyMaterialService,
		subTopicID:           subTopicID,
		router:               router,
		sub:                  sub,
		logger:               logger,
	}
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
	decodedMsg, err := study_material_models.DecodeToAttachmentToParse(msg.Payload)
	if err != nil {
		s.logger.Err(err).Msg("failed to decode message")
		return nil // need to send ack to watermill
	}
	s.logger.Info().Interface("decodedMsg", decodedMsg).Msg("decoded message")
	err = s.studyMaterialService.ProcessAttachmentToParse(msg.Context(), decodedMsg)
	if err != nil {
		s.logger.Err(err).Msg("failed to ProcessAttachmentToParse")
		return nil // need to send ack to watermill
	}
	return nil
}
