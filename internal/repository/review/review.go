package review_repo

import (
	"context"
	"errors"
	custom_errors "github.com/Petr09Mitin/xrust-beze-back/internal/models/error"
	user_model "github.com/Petr09Mitin/xrust-beze-back/internal/models/user"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ReviewRepo interface {
	Create(ctx context.Context, review *user_model.Review) (*user_model.Review, error)
	GetByUserIDByAndUserIDTo(ctx context.Context, userIDBy string, UserIDTo string) (*user_model.Review, error)
}

type ReviewRepoImpl struct {
	mongoDB *mongo.Collection
	logger  zerolog.Logger
}

func NewReviewRepo(mongoDB *mongo.Collection, logger zerolog.Logger) ReviewRepo {
	return &ReviewRepoImpl{
		mongoDB: mongoDB,
		logger:  logger,
	}
}

func (r *ReviewRepoImpl) Create(ctx context.Context, review *user_model.Review) (*user_model.Review, error) {
	result, err := r.mongoDB.InsertOne(ctx, review)
	if err != nil {
		r.logger.Error().Err(err).Msg("failed to create review in MongoDB")
		return nil, err
	}
	review.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return review, nil
}

func (r *ReviewRepoImpl) GetByUserIDByAndUserIDTo(ctx context.Context, userIDBy string, userIDTo string) (*user_model.Review, error) {
	res := r.mongoDB.FindOne(ctx, bson.M{"user_id_by": userIDBy, "user_id_to": userIDTo})
	review := &user_model.Review{}
	err := res.Decode(review)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return &user_model.Review{}, custom_errors.ErrNotFound
		}
		return nil, err
	}
	return review, nil
}
