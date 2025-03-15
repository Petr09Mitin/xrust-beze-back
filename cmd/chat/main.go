package main

import (
	"github.com/Petr09Mitin/xrust-beze-back/internal/router"
	chat_service "github.com/Petr09Mitin/xrust-beze-back/internal/services/chat"

	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	chatService := chat_service.NewChatService()
	c := router.NewChat(chatService)
	err := c.Start()
	if err != nil {
		panic(err)
	}
	client, err := connectToMongo()
	if err != nil {
		log.Fatal("Ошибка подключения к MongoDB:", err)
	}
	defer client.Disconnect(context.Background())

	err = insertMessage(client, "Привет, MongoDB!")
	if err != nil {
		log.Fatal("Ошибка вставки сообщения:", err)
	}
}

func connectToMongo() (*mongo.Client, error) {
	// mongoURI := "mongodb://root:root@mongo_db:27017/admin"
	mongoURI := "mongodb://admin:admin@mongo_db:27017/xrust_beze?authSource=admin"

	clientOptions := options.Client().ApplyURI(mongoURI)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	fmt.Println("Подключено к MongoDB")
	return client, nil
}

func insertMessage(client *mongo.Client, message string) error {
	collection := client.Database("xrust_beze").Collection("chats")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	doc := bson.M{"message": message, "timestamp": time.Now()}

	_, err := collection.InsertOne(ctx, doc)
	if err != nil {
		return err
	}

	fmt.Println("Сообщение сохранено в MongoDB:", message)
	return nil
}
