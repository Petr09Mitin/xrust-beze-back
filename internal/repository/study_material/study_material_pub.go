package study_material_repo

import (
	"context"
	study_material_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/study_material"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/rs/zerolog"
)

type StudyMaterialPub interface {
	PublishAttachmentToParse(ctx context.Context, attachment *study_material_models.AttachmentToParse) error
}

type StudyMaterialPubImpl struct {
	pubTopicID string
	pub        message.Publisher
	logger     zerolog.Logger
}

func NewStudyMaterialPub(pubTopicID string, pub message.Publisher, logger zerolog.Logger) StudyMaterialPub {
	return &StudyMaterialPubImpl{
		pubTopicID: pubTopicID,
		pub:        pub,
		logger:     logger,
	}
}

func (s *StudyMaterialPubImpl) PublishAttachmentToParse(_ context.Context, attachment *study_material_models.AttachmentToParse) error {
	return s.pub.Publish(s.pubTopicID, message.NewMessage(
		watermill.NewUUID(),
		attachment.Encode(),
	))
}
