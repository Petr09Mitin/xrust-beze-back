package study_material_repo

import (
	"context"
	study_material_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/study_material"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type StudyMaterialRepo interface {
	Insert(ctx context.Context, material *study_material_models.StudyMaterial) (*study_material_models.StudyMaterial, error)
	GetByTag(ctx context.Context, tag string) ([]*study_material_models.StudyMaterial, error)
	GetByAuthorID(ctx context.Context, authorID string) ([]*study_material_models.StudyMaterial, error)
	Delete(ctx context.Context, materialID string) error
}

type StudyMaterialRepoImpl struct {
	mongoDB *mongo.Collection
	logger  zerolog.Logger
}

func NewStudyMaterialRepo(mongoDB *mongo.Collection, logger zerolog.Logger) StudyMaterialRepo {
	return &StudyMaterialRepoImpl{
		mongoDB: mongoDB,
		logger:  logger,
	}
}

func (s *StudyMaterialRepoImpl) Insert(ctx context.Context, material *study_material_models.StudyMaterial) (*study_material_models.StudyMaterial, error) {
	return nil, nil
}

func (s *StudyMaterialRepoImpl) Delete(ctx context.Context, materialID string) error {
	return nil
}

func (s *StudyMaterialRepoImpl) GetByTag(ctx context.Context, tag string) ([]*study_material_models.StudyMaterial, error) {
	return nil, nil
}

func (s *StudyMaterialRepoImpl) GetByAuthorID(ctx context.Context, authorID string) ([]*study_material_models.StudyMaterial, error) {
	return nil, nil
}
