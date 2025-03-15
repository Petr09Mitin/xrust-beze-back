package mongodb

import (
	"context"
	"time"

	"github.com/yourusername/chatapp/internal/domain/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MessageRepository struct {
	collection *mongo.Collection
}

func NewMessageRepository(client *mongo.Client, dbName, collectionName string) *MessageRepository {
	collection := client.Database(dbName).Collection(collectionName)
	return &MessageRepository{
		collection: collection,
	}
}

// сохраняет новое сообщение в базе данных
func (r *MessageRepository) Create(ctx context.Context, message *models.Message) error {
	message.Timestamp = time.Now()
	
	_, err := r.collection.InsertOne(ctx, message)
	if err != nil {
		return err
	}
	
	return nil
}

// возвращает все сообщения
func (r *MessageRepository) FindAll(ctx context.Context) ([]*models.Message, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var messages []*models.Message
	if err = cursor.All(ctx, &messages); err != nil {
		return nil, err
	}
	
	return messages, nil
}

// находит сообщение по ID
func (r *MessageRepository) FindByID(ctx context.Context, id string) (*models.Message, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	
	var message models.Message
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&message)
	if err != nil {
		return nil, err
	}
	
	return &message, nil
}
