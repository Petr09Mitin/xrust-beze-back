package study_material

import (
	"context"
	study_material_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/study_material"
	study_material_repo "github.com/Petr09Mitin/xrust-beze-back/internal/repository/study_material"
	"github.com/rs/zerolog"
)

type StudyMaterialService interface {
	ProcessAttachmentToParse(ctx context.Context, attachment *study_material_models.AttachmentToParse) error
}

type StudyMaterialServiceImpl struct {
	studyMaterialRepo study_material_repo.StudyMaterialRepo
	logger            zerolog.Logger
}

func NewStudyMaterialService(studyMaterialRepo study_material_repo.StudyMaterialRepo, logger zerolog.Logger) StudyMaterialService {
	return &StudyMaterialServiceImpl{
		studyMaterialRepo: studyMaterialRepo,
		logger:            logger,
	}
}

func (s *StudyMaterialServiceImpl) ProcessAttachmentToParse(ctx context.Context, attachment *study_material_models.AttachmentToParse) error {
	return nil
}
