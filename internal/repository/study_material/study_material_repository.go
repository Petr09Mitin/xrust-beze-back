package study_material_repo

import (
	"context"
	"errors"
	"fmt"

	custom_errors "github.com/Petr09Mitin/xrust-beze-back/internal/models/error"
	study_material_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/study_material"
	"github.com/rs/zerolog"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
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
	var bsonMaterial study_material_models.BSONStudyMaterial
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, custom_errors.ErrInvalidIDType
	}
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&bsonMaterial)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, custom_errors.ErrNotFound
		}
		return nil, err
	}

	return bsonMaterial.ToStudyMaterial(), nil
}

// func (r *studyMaterialAPIRepository) GetByTags(ctx context.Context, tags []string) ([]*study_material_models.StudyMaterial, error) {
// 	// filter := bson.M{"tags": tag}
// 	filter := bson.M{"tags": bson.M{"$in": tags}}
// 	return r.find(ctx, filter)
// }

func (r *studyMaterialAPIRepository) GetByTags(ctx context.Context, tags []string) ([]*study_material_models.StudyMaterial, error) {
	var regexes []interface{}
	for _, tag := range tags {
		regexes = append(regexes, bson.M{"$regex": fmt.Sprintf("^%s$", tag), "$options": "i"})
	}
	filter := bson.M{"tags": bson.M{"$in": regexes}}
	return r.find(ctx, filter)
}

func (r *studyMaterialAPIRepository) GetByName(ctx context.Context, name string) ([]*study_material_models.StudyMaterial, error) {
	// Ищутся все материалы, содержащие в названии строку name
	// Регистр не учитывается
	filter := bson.M{"name": bson.M{"$regex": name, "$options": "i"}}
	return r.find(ctx, filter)
}

func (r *studyMaterialAPIRepository) GetByAuthorID(ctx context.Context, authorID string) ([]*study_material_models.StudyMaterial, error) {
	// objectAuthorID, err := primitive.ObjectIDFromHex(authorID)
	// if err != nil {
	// 	return nil, err
	// }
	filter := bson.M{"author_id": authorID}
	return r.find(ctx, filter)
}

func (r *studyMaterialAPIRepository) Delete(ctx context.Context, materialID string) error {
	objectMaterialID, err := bson.ObjectIDFromHex(materialID)
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
	defer func() {
		err = cursor.Close(ctx)
		if err != nil {
			r.logger.Error().Err(err).Msg("failed to close cursor")
		}
	}()

	materials := make([]*study_material_models.StudyMaterial, 0, cursor.RemainingBatchLength())
	for cursor.Next(ctx) {
		curr := study_material_models.BSONStudyMaterial{}
		err = cursor.Decode(&curr)
		if err != nil {
			return nil, err
		}
		materials = append(materials, curr.ToStudyMaterial())
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return materials, nil
}
