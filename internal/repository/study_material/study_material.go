package study_material_repo

import (
	"context"
	study_material_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/study_material"
)

type StudyMaterialRepo interface {
	Insert(ctx context.Context, material *study_material_models.StudyMaterial) (*study_material_models.StudyMaterial, error)
	GetByTag(ctx context.Context, tag string) ([]*study_material_models.StudyMaterial, error)
	GetByAuthorID(ctx context.Context, authorID string) ([]*study_material_models.StudyMaterial, error)
	Delete(ctx context.Context, materialID string) error
}
