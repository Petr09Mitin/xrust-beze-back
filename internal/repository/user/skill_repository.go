package user_repo

import (
	"context"
	"github.com/rs/zerolog"
	"time"

	skill_model "github.com/Petr09Mitin/xrust-beze-back/internal/models/user"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type SkillRepo interface {
	GetAllSkills(ctx context.Context) ([]skill_model.SkillsByCategory, error)
	GetSkillsByCategory(ctx context.Context, category string) ([]string, error)
	GetAllCategories(ctx context.Context) ([]string, error)
}

type skillRepository struct {
	collection *mongo.Collection
	timeout    time.Duration
	logger     zerolog.Logger
}

func NewSkillRepository(db *mongo.Database, timeout time.Duration, logger zerolog.Logger) SkillRepo {
	return &skillRepository{
		collection: db.Collection("skills"),
		timeout:    timeout,
		logger:     logger,
	}
}

func (r *skillRepository) GetAllSkills(ctx context.Context) ([]skill_model.SkillsByCategory, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer func() {
		err = cursor.Close(ctx)
		if err != nil {
			r.logger.Error().Err(err).Msg("failed to close cursor")
		}
	}()

	var skills []skill_model.SkillsByCategory
	if err := cursor.All(ctx, &skills); err != nil {
		return nil, err
	}

	return skills, nil
}

func (r *skillRepository) GetSkillsByCategory(ctx context.Context, category string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var result skill_model.SkillsByCategory
	err := r.collection.FindOne(ctx, bson.M{"category": category}).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result.Skills, nil
}

func (r *skillRepository) GetAllCategories(ctx context.Context) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	// Используем Distinct для получения уникальных категорий
	res := r.collection.Distinct(ctx, "category", bson.M{})
	if err := res.Err(); err != nil {
		return nil, err
	}
	result := make([]string, 0)
	err := res.Decode(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
