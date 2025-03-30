package channelrepo

import (
	"context"
	"errors"
	"fmt"
	chat_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/chat"
	custom_errors "github.com/Petr09Mitin/xrust-beze-back/internal/models/error"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"slices"
)

type ChannelRepository interface {
	InsertChannel(ctx context.Context, channel chat_models.Channel) (chat_models.Channel, error)
	GetChannelByID(ctx context.Context, id string) (chat_models.Channel, error)
	GetChannelsByUserID(ctx context.Context, userID string, limit, offset int64) ([]chat_models.Channel, error)
	GetByUserIDs(ctx context.Context, userIDs []string) (chat_models.Channel, error)
}

type ChannelRepositoryImpl struct {
	mongoDB *mongo.Collection
}

func NewChannelRepository(mongoDB *mongo.Collection) ChannelRepository {
	return &ChannelRepositoryImpl{
		mongoDB: mongoDB,
	}
}

func (r *ChannelRepositoryImpl) InsertChannel(ctx context.Context, channel chat_models.Channel) (chat_models.Channel, error) {
	// sort userIDs for speeding up the search by user_ids, as mongo stores arrays in stable order
	slices.Sort(channel.UserIDs)
	res, err := r.mongoDB.InsertOne(ctx, channel)
	if err != nil {
		return channel, err
	}

	channel.ID = res.InsertedID.(bson.ObjectID).Hex()

	return channel, nil
}

func (r *ChannelRepositoryImpl) GetChannelByID(ctx context.Context, id string) (chat_models.Channel, error) {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return chat_models.Channel{}, err
	}
	res := r.mongoDB.FindOne(ctx, bson.M{
		"_id": objID,
	})
	curr := chat_models.BSONChannel{}
	err = res.Decode(&curr)
	if err != nil {
		return chat_models.Channel{}, err
	}
	channel := curr.ToChannel()

	return channel, nil
}

func (r *ChannelRepositoryImpl) GetChannelsByUserID(ctx context.Context, userID string, limit, offset int64) ([]chat_models.Channel, error) {
	cur, err := r.mongoDB.Find(
		ctx,
		bson.M{
			"user_ids": userID,
		},
		options.Find().SetSort(
			bson.M{
				"created_at": -1,
			},
		).SetLimit(limit).SetSkip(offset),
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = cur.Close(ctx)
		if err != nil {
			fmt.Println(err)
			return
		}
	}()
	res := make([]chat_models.Channel, 0, cur.RemainingBatchLength())
	for cur.Next(ctx) {
		curr := chat_models.BSONChannel{}
		err = cur.Decode(&curr)
		if err != nil {
			return nil, err
		}
		res = append(res, curr.ToChannel())
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (r *ChannelRepositoryImpl) GetByUserIDs(ctx context.Context, userIDs []string) (chat_models.Channel, error) {
	// userIDs must be sorted to perform order-independent search
	slices.Sort(userIDs)
	res := r.mongoDB.FindOne(ctx, bson.M{
		"user_ids": bson.M{
			"$eq": userIDs,
		},
	})
	curr := chat_models.BSONChannel{}
	err := res.Decode(&curr)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return chat_models.Channel{}, custom_errors.ErrNotFound
		}
		return chat_models.Channel{}, err
	}
	channel := curr.ToChannel()

	return channel, nil
}
