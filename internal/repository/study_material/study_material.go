package study_material_repo

import (
	"context"

	study_material_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/study_material"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type StudyMaterialRepo interface {
	Insert(ctx context.Context, material *study_material_models.StudyMaterial) (*study_material_models.StudyMaterial, error)
}

type studyMaterialRepository struct {
	mongoDB *mongo.Collection
	logger  zerolog.Logger
}

func NewStudyMaterialRepo(mongoDB *mongo.Collection, logger zerolog.Logger) StudyMaterialRepo {
	return &studyMaterialRepository{
		mongoDB: mongoDB,
		logger:  logger,
	}
}

func (s *studyMaterialRepository) Insert(ctx context.Context, material *study_material_models.StudyMaterial) (*study_material_models.StudyMaterial, error) {
	res, err := s.mongoDB.InsertOne(ctx, material)
	if err != nil {
		return nil, err
	}
	material.ID = res.InsertedID.(bson.ObjectID).Hex()
	return material, nil
}
