package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SkillDocument struct {
	Category string   `bson:"category"`
	Skills   []string `bson:"skills"`
}

func main() {
	log.Println("Connecting to MongoDB...")
	uri := "mongodb://admin:admin@localhost:27017/xrust_beze?authSource=admin"
	dbName := "xrust_beze"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}
	log.Println("Connected to MongoDB successfully")

	db := client.Database(dbName)
	collection := db.Collection("skills")

	// Чтение файла
	data, err := os.ReadFile("scripts/skills_by_category.json")
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	// Парсинг JSON
	var skills []SkillDocument
	if err := json.Unmarshal(data, &skills); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	// Загрузка в MongoDB
	for _, skill := range skills {
		filter := bson.M{"category": skill.Category}
		update := bson.M{"$setOnInsert": skill}
		_, err := collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))

		if err != nil {
			log.Printf("Failed to upsert category %s: %v", skill.Category, err)
			continue
		}
	}
}
