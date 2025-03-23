package mongodb

import (
	"context"
	"errors"
	"time"

	user_model "github.com/Petr09Mitin/xrust-beze-back/internal/models/user"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type userRepository struct {
	collection *mongo.Collection
	timeout    time.Duration
}

// NewUserRepository создает новый репозиторий пользователей
func NewUserRepository(db *mongo.Database, timeout time.Duration) user_model.Repository {
	return &userRepository{
		collection: db.Collection("users"),
		timeout:    timeout,
	}
}

// Create создает нового пользователя
func (r *userRepository) Create(user *user_model.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.LastActiveAt = time.Now()

	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return err
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetByID получает пользователя по ID
func (r *userRepository) GetByID(id string) (*user_model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var u user_model.User
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&u)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &u, nil
}

// GetByEmail получает пользователя по email
func (r *userRepository) GetByEmail(email string) (*user_model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	var u user_model.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&u)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &u, nil
}

// GetByUsername получает пользователя по имени пользователя
func (r *userRepository) GetByUsername(username string) (*user_model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	var u user_model.User
	err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&u)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &u, nil
}

// Update обновляет пользователя
func (r *userRepository) Update(user *user_model.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	user.UpdatedAt = time.Now()

	_, err := r.collection.ReplaceOne(
		ctx,
		bson.M{"_id": user.ID},
		user,
	)
	return err
}

// Delete удаляет пользователя
func (r *userRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}

// List возвращает список пользователей с пагинацией
func (r *userRepository) List(page, limit int) ([]*user_model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	skip := (page - 1) * limit

	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*user_model.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// FindBySkills находит пользователей по навыкам
func (r *userRepository) FindBySkills(skillsToLearn, skillsToShare []string) ([]*user_model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	filter := bson.M{}
	if len(skillsToLearn) > 0 {
		filter["skills_to_share.name"] = bson.M{"$in": skillsToLearn}
	}
	if len(skillsToShare) > 0 {
		filter["skills_to_learn.name"] = bson.M{"$in": skillsToShare}
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*user_model.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
} 