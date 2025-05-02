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
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ReviewRepo interface {
	Create(ctx context.Context, review *user_model.Review) (*user_model.Review, error)
	GetByUserIDByAndUserIDTo(ctx context.Context, userIDBy string, UserIDTo string) (*user_model.Review, error)
	GetReviewsByUserIDTo(ctx context.Context, userIDTo string) ([]*user_model.Review, error)
	GetAvgRatingsByUserIDs(ctx context.Context, userIDs []string) (map[string]float64, error)
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
	res := r.mongoDB.FindOne(ctx, bson.M{"user_id_to": userIDTo, "user_id_by": userIDBy})
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

func (r *ReviewRepoImpl) GetReviewsByUserIDTo(ctx context.Context, userIDTo string) ([]*user_model.Review, error) {
	cur, err := r.mongoDB.Find(
		ctx,
		bson.M{
			"user_id_to": userIDTo,
		},
		options.Find().SetSort(
			bson.M{
				"created": -1,
			},
		),
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = cur.Close(ctx)
		if err != nil {
			r.logger.Err(err)
			return
		}
	}()
	res := make([]*user_model.Review, 0, cur.RemainingBatchLength())
	for cur.Next(ctx) {
		curr := &user_model.BSONReview{}
		err = cur.Decode(&curr)
		if err != nil {
			return nil, err
		}
		res = append(res, curr.ToReview())
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (r *ReviewRepoImpl) GetAvgRatingsByUserIDs(ctx context.Context, userIDs []string) (map[string]float64, error) {
	cur, err := r.mongoDB.Aggregate(
		ctx,
		mongo.Pipeline{
			bson.D{
				{"$group", bson.D{
					{"_id", "$user_id_to"},
					{
						"avg_rating",
						bson.D{
							{"$avg", "$rating"},
						},
					},
				},
				},
			},
			bson.D{
				{"$match", bson.D{
					{"_id", bson.D{
						{"$in", userIDs},
					}},
				}},
			},
		},
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = cur.Close(ctx)
		if err != nil {
			r.logger.Err(err)
			return
		}
	}()
	res := make([]*user_model.BSONAvgRating, 0, cur.RemainingBatchLength())
	for cur.Next(ctx) {
		curr := &user_model.BSONAvgRating{}
		err = cur.Decode(&curr)
		if err != nil {
			return nil, err
		}
		res = append(res, curr)
		r.logger.Info().Any("res", curr)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	ratingsMap := make(map[string]float64, len(res))
	for _, rating := range res {
		ratingsMap[rating.UserID] = rating.Rating
	}
	return ratingsMap, nil
}
