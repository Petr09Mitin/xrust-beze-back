package message_repo

import (
	"context"
	chat_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/chat"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	MessagePubTopic = "xb.msg.pub"
)

type MessageRepo interface {
	GetMessagesByChannelID(ctx context.Context, channelID string, limit, offset int64) ([]chat_models.Message, error)
	GetMessageByID(ctx context.Context, id string) (*chat_models.Message, error)
	InsertMessage(ctx context.Context, msg chat_models.Message) (chat_models.Message, error)
	UpdateMessage(ctx context.Context, msg chat_models.Message) error
	DeleteMessage(ctx context.Context, msg chat_models.Message) error
	PublishMessage(ctx context.Context, msg chat_models.Message) error
}

type MessageRepoImpl struct {
	p       message.Publisher
	mongoDB *mongo.Collection
	logger  zerolog.Logger
}

func NewMessageRepo(p message.Publisher, mongoDB *mongo.Collection, logger zerolog.Logger) MessageRepo {
	return &MessageRepoImpl{
		p:       p,
		mongoDB: mongoDB,
		logger:  logger,
	}
}

func (m *MessageRepoImpl) GetMessageByID(ctx context.Context, id string) (*chat_models.Message, error) {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	res := m.mongoDB.FindOne(ctx, bson.M{
		"_id": objID,
	})
	msg := &chat_models.Message{}
	err = res.Decode(msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (m *MessageRepoImpl) InsertMessage(ctx context.Context, msg chat_models.Message) (chat_models.Message, error) {
	res, err := m.mongoDB.InsertOne(ctx, msg)
	if err != nil {
		return msg, err
	}

	msg.MessageID = res.InsertedID.(bson.ObjectID).Hex()

	return msg, nil
}

func (m *MessageRepoImpl) UpdateMessage(ctx context.Context, msg chat_models.Message) error {
	objID, err := bson.ObjectIDFromHex(msg.MessageID)
	if err != nil {
		return err
	}
	_, err = m.mongoDB.UpdateByID(ctx, objID, m.getUpdateDocumentFromMsg(msg))
	if err != nil {
		return err
	}

	return nil
}

func (m *MessageRepoImpl) DeleteMessage(ctx context.Context, msg chat_models.Message) error {
	objID, err := bson.ObjectIDFromHex(msg.MessageID)
	if err != nil {
		return err
	}
	filter := bson.D{
		{"_id", objID},
	}
	_, err = m.mongoDB.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	return nil
}

func (m *MessageRepoImpl) PublishMessage(_ context.Context, msg chat_models.Message) error {
	return m.p.Publish(MessagePubTopic, message.NewMessage(
		watermill.NewUUID(),
		msg.Encode(),
	))
}

func (m *MessageRepoImpl) getUpdateDocumentFromMsg(msg chat_models.Message) bson.M {
	return bson.M{
		"$set": bson.M{
			"payload":    msg.Payload,
			"updated_at": msg.UpdatedAt,
		},
	}
}

func (m *MessageRepoImpl) GetMessagesByChannelID(ctx context.Context, channelID string, limit, offset int64) ([]chat_models.Message, error) {
	cur, err := m.mongoDB.Find(
		ctx,
		bson.M{
			"channel_id": channelID,
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
			m.logger.Err(err)
			return
		}
	}()
	res := make([]chat_models.Message, 0, cur.RemainingBatchLength())
	for cur.Next(ctx) {
		curr := chat_models.BSONMessage{}
		err = cur.Decode(&curr)
		if err != nil {
			return nil, err
		}
		res = append(res, curr.ToMessage())
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	return res, nil
}
