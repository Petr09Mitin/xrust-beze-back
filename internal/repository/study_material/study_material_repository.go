package study_material_repo

import (
	"context"

	custom_errors "github.com/Petr09Mitin/xrust-beze-back/internal/models/error"
	study_material_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/study_material"
	"github.com/rs/zerolog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type StudyMaterialAPIRepository interface {
	GetByID(ctx context.Context, id string) (*study_material_models.StudyMaterial, error)
	GetByTags(ctx context.Context, tag []string) ([]*study_material_models.StudyMaterial, error)
	GetByAuthorID(ctx context.Context, authorID string) ([]*study_material_models.StudyMaterial, error)
	Delete(ctx context.Context, materialID string) error
	GetByName(ctx context.Context, name string) ([]*study_material_models.StudyMaterial, error)
}

type studyMaterialAPIRepository struct {
	collection *mongo.Collection
	logger     zerolog.Logger
}

func NewStudyMaterialAPIRepository(db *mongo.Database, logger zerolog.Logger) StudyMaterialAPIRepository {
	return &studyMaterialAPIRepository{
		collection: db.Collection("study_materials"),
		logger:     logger,
	}
}

func (r *studyMaterialAPIRepository) GetByID(ctx context.Context, id string) (*study_material_models.StudyMaterial, error) {
	var material study_material_models.StudyMaterial
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&material)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, custom_errors.ErrNotFound
		}
		return nil, err
	}

	return &material, nil
}

func (r *studyMaterialAPIRepository) GetByTags(ctx context.Context, tags []string) ([]*study_material_models.StudyMaterial, error) {
	// filter := bson.M{"tags": tag}
	filter := bson.M{"tags": bson.M{"$in": tags}}
	return r.find(ctx, filter)
}

func (r *studyMaterialAPIRepository) GetByName(ctx context.Context, name string) ([]*study_material_models.StudyMaterial, error) {
	// Ищутся все материалы, содержащие в названии строку name
	// Регистр не учитывается
	filter := bson.M{"name": bson.M{"$regex": name, "$options": "i"}}
	return r.find(ctx, filter)
}

func (r *studyMaterialAPIRepository) GetByAuthorID(ctx context.Context, authorID string) ([]*study_material_models.StudyMaterial, error) {
	objectAuthorID, err := primitive.ObjectIDFromHex(authorID)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"author_id": objectAuthorID}
	return r.find(ctx, filter)
}

func (r *studyMaterialAPIRepository) Delete(ctx context.Context, materialID string) error {
	objectMaterialID, err := primitive.ObjectIDFromHex(materialID)
	if err != nil {
		return err
	}
	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectMaterialID})
	return err
}

func (r *studyMaterialAPIRepository) find(ctx context.Context, filter interface{}) ([]*study_material_models.StudyMaterial, error) {
	findOptions := options.Find().SetSort(bson.D{bson.E{Key: "created", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var materials []*study_material_models.StudyMaterial
	if err := cursor.All(ctx, &materials); err != nil {
		return nil, err
	}

	return materials, nil
}
