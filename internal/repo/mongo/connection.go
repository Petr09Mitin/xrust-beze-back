package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Config содержит настройки подключения к MongoDB
type Config struct {
	URI        string
	Database   string
	AuthSource string
	Username   string
	Password   string
}

// NewConnection создает новое подключение к MongoDB
func NewConnection(cfg Config) (*mongo.Client, error) {
	mongoURI := fmt.Sprintf("mongodb://%s:%s@%s/%s?authSource=%s",
		cfg.Username, cfg.Password, cfg.URI, cfg.Database, cfg.AuthSource)

	clientOptions := options.Client().ApplyURI(mongoURI)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Проверяем подключение
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	fmt.Println("~!!!~ Подключено к MongoDB")
	return client, nil
}