package message_repo

import (
	"context"
	chat_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/chat"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const (
	MessagePubTopic = "xb.msg.pub"
)

type MessageRepo interface {
	InsertMessage(ctx context.Context, msg chat_models.Message) (chat_models.Message, error)
	UpdateMessage(ctx context.Context, msg chat_models.Message) error
	DeleteMessage(ctx context.Context, msg chat_models.Message) error
	PublishMessage(ctx context.Context, msg chat_models.Message) error
}

type MessageRepoImpl struct {
	p       message.Publisher
	mongoDB *mongo.Collection
}

func NewMessageRepo(p message.Publisher, mongoDB *mongo.Collection) MessageRepo {
	return &MessageRepoImpl{
		p:       p,
		mongoDB: mongoDB,
	}
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
