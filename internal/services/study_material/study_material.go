package study_material_service

import (
	"context"
	"strings"
	"time"

	study_material_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/study_material"
	study_material_repo "github.com/Petr09Mitin/xrust-beze-back/internal/repository/study_material"
	"github.com/rs/zerolog"
)

type StudyMaterialService interface {
	ProcessAttachmentToParse(ctx context.Context, attachment *study_material_models.AttachmentToParse) error
}

type StudyMaterialServiceImpl struct {
	studyMaterialRepo study_material_repo.StudyMaterialRepo
	mlTaggerRepo      study_material_repo.MLTaggerRepo
	fileRepo          study_material_repo.FileRepo
	logger            zerolog.Logger
}

func NewStudyMaterialService(studyMaterialRepo study_material_repo.StudyMaterialRepo, mlTaggerRepo study_material_repo.MLTaggerRepo, fileRepo study_material_repo.FileRepo, logger zerolog.Logger) StudyMaterialService {
	return &StudyMaterialServiceImpl{
		studyMaterialRepo: studyMaterialRepo,
		mlTaggerRepo:      mlTaggerRepo,
		fileRepo:          fileRepo,
		logger:            logger,
	}
}

func (s *StudyMaterialServiceImpl) ProcessAttachmentToParse(ctx context.Context, attachment *study_material_models.AttachmentToParse) error {
	if !strings.HasSuffix(attachment.Filename, ".pdf") && !strings.HasSuffix(attachment.Filename, ".docx") {
		s.logger.Info().Interface("attachment", attachment).Msg("attachment is not a pdf file")
		return nil
	}

	res, err := s.mlTaggerRepo.ProcessAttachment(ctx, attachment)
	if err != nil {
		return err
	}
	if !res.IsStudyMaterial || res.StudyMaterial == nil {
		s.logger.Info().Interface("attachment", attachment).Msg("attachment is not a study material")
		return nil
	}
	attachment.Filename, err = s.fileRepo.CopyAttachmentToStudyFiles(ctx, attachment.Filename)
	if err != nil {
		return err
	}
	createdAt := time.Now().Unix()
	material, err := s.studyMaterialRepo.Insert(ctx, &study_material_models.StudyMaterial{
		Name:     res.StudyMaterial.Name,
		Filename: res.StudyMaterial.Filename,
		Tags:     res.StudyMaterial.Tags,
		AuthorID: attachment.AuthorID,
		Created:  createdAt,
		Updated:  createdAt,
	})
	if err != nil {
		return err
	}
	s.logger.Debug().Interface("material", material).Msg("study material created")
	return nil
}
