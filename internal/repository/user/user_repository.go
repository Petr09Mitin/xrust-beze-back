package user_repo

import (
	"context"
	"errors"
	custom_errors "github.com/Petr09Mitin/xrust-beze-back/internal/models/error"
	"github.com/rs/zerolog"
	"time"

	user_model "github.com/Petr09Mitin/xrust-beze-back/internal/models/user"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepo interface {
	Create(ctx context.Context, user *user_model.User) error
	GetByID(ctx context.Context, id string) (*user_model.User, error)
	GetByEmail(ctx context.Context, email string) (*user_model.User, error)
	GetByUsername(ctx context.Context, username string) (*user_model.User, error)
	Update(ctx context.Context, user *user_model.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, page, limit int) ([]*user_model.User, error)
	FindBySkills(ctx context.Context, skillsToLearn []string) ([]*user_model.User, error)
	FindByUsername(ctx context.Context, name string, limit, offset int64) ([]*user_model.User, error)
}

type userRepository struct {
	collection *mongo.Collection
	timeout    time.Duration
	logger     zerolog.Logger
}

func NewUserRepository(db *mongo.Database, timeout time.Duration, logger zerolog.Logger) UserRepo {
	return &userRepository{
		collection: db.Collection("users"),
		timeout:    timeout,
		logger:     logger,
	}
}

func (r *userRepository) Create(ctx context.Context, user *user_model.User) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
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

func (r *userRepository) GetByID(ctx context.Context, id string) (*user_model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var u user_model.User
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&u)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, custom_errors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*user_model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var u user_model.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&u)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, custom_errors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*user_model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var u user_model.User
	err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&u)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, custom_errors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *userRepository) Update(ctx context.Context, user *user_model.User) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	user.UpdatedAt = time.Now()

	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": user.ID}, user)

	return err
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}

func (r *userRepository) List(ctx context.Context, page, limit int) ([]*user_model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	skip := (page - 1) * limit

	opts := options.Find().SetSkip(int64(skip)).SetLimit(int64(limit))
	opts.SetSort(bson.M{"created_at": -1}) // Сначала новые
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = cursor.Close(ctx)
		if err != nil {
			r.logger.Error().Err(err).Msg("failed to close cursor")
		}
	}()

	var users []*user_model.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userRepository) FindBySkills(ctx context.Context, skillsToLearn []string) ([]*user_model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	filter := bson.M{}
	if len(skillsToLearn) > 0 {
		filter["skills_to_share.name"] = bson.M{"$in": skillsToLearn}
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = cursor.Close(ctx)
		if err != nil {
			r.logger.Error().Err(err).Msg("failed to close cursor")
		}
	}()

	var users []*user_model.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userRepository) FindByUsername(ctx context.Context, name string, limit, offset int64) ([]*user_model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	filter := mongo.Pipeline{
		{{
			"$match", bson.D{
				{"$text", bson.D{
					{"$search", name},
				}}},
		}},
		{{
			"$project", bson.M{
				"username": 1,
				"score": bson.D{
					{"$meta", "textScore"},
				},
			},
		}},
		{{
			"$sort", bson.D{
				{"score", 1},
			},
		}},
		{{"$limit", limit + offset}},
		{{"$skip", offset}},
	}
	cursor, err := r.collection.Aggregate(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = cursor.Close(ctx)
		if err != nil {
			r.logger.Error().Err(err).Msg("failed to close cursor")
		}
	}()

	users := make([]*user_model.User, 0)
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}
