package user_repo

import (
	"context"
	"time"

	skill_model "github.com/Petr09Mitin/xrust-beze-back/internal/models/user"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type SkillRepo interface {
	GetAllSkills(ctx context.Context) ([]skill_model.SkillsByCategory, error)
	GetSkillsByCategory(ctx context.Context, category string) ([]string, error)
	GetAllCategories(ctx context.Context) ([]string, error)
}

type skillRepository struct {
	collection *mongo.Collection
	timeout    time.Duration
}

func NewSkillRepository(db *mongo.Database, timeout time.Duration) SkillRepo {
	return &skillRepository{
		collection: db.Collection("skills"),
		timeout:    timeout,
	}
}

func (r *skillRepository) GetAllSkills(ctx context.Context) ([]skill_model.SkillsByCategory, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

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
	categories, err := r.collection.Distinct(ctx, "category", bson.M{})
	if err != nil {
		return nil, err
	}

	// Конвертируем interface{} в []string
	var result []string
	for _, cat := range categories {
		if str, ok := cat.(string); ok {
			result = append(result, str)
		}
	}

	return result, nil
}
