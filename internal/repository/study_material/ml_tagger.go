package study_material_repo

import (
	"context"
	study_material_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/study_material"
)

type MLTaggerRepo interface {
	ProcessAttachment(ctx context.Context, attachment *study_material_models.AttachmentToParse) (*study_material_models.ParsedAttachmentResponse, error)
}

type MLTaggerRepoImpl struct {
}

func NewMLTaggerRepo() MLTaggerRepo {
	return &MLTaggerRepoImpl{}
}

func (m *MLTaggerRepoImpl) ProcessAttachment(ctx context.Context, attachment *study_material_models.AttachmentToParse) (*study_material_models.ParsedAttachmentResponse, error) {
	// TODO: add business logic
	return &study_material_models.ParsedAttachmentResponse{
		IsStudyMaterial: true,
		StudyMaterial: &study_material_models.StudyMaterial{
			Name:     "Figma Study Material",
			Filename: attachment.Filename,
			AuthorID: attachment.AuthorID,
			Tags:     []string{"Figma"},
		},
	}, nil
}
