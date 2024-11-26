// config/db.go

package config

import (
    "context"
    "log"
    "os"
    "time"
    //"fmt"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
)

var Client *mongo.Client

func ConnectDB() {
    // Get MongoDB URI from environment variable or use default
    mongoURI := os.Getenv("MONGO_URI")
    if mongoURI == "" {
        // Default local MongoDB URI if not set
        mongoURI = "mongodb://root:example@localhost:27017/onboarding?authSource=admin"
    }

    // Set client options
    clientOptions := options.Client().ApplyURI(mongoURI)

    // Connect to MongoDB
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Create client
    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        log.Fatal("Error creating MongoDB client:", err)
    }

    // Ping the database
    err = client.Ping(ctx, nil)
    if err != nil {
        log.Fatal("Error connecting to MongoDB:", err)
    }

    Client = client
    log.Println("Connected to MongoDB!")

    // Initialize indexes
    initializeIndexes(client)
}

// initializeIndexes creates necessary indexes for the collections
func initializeIndexes(client *mongo.Client) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Get the submissions collection
    collection := client.Database("onboarding").Collection("submissions")

    // Create indexes
    indexes := []mongo.IndexModel{
        {
            Keys: bson.D{{Key: "issuer_info.company_name", Value: 1}},
            Options: options.Index().SetUnique(true),
        },
        {
            Keys: bson.D{{Key: "status", Value: 1}},
        },
        {
            Keys: bson.D{{Key: "submitted_at", Value: -1}},
        },
    }

    // Create the indexes
    _, err := collection.Indexes().CreateMany(ctx, indexes)
    if err != nil {
        log.Printf("Warning: Error creating indexes: %v", err)
    }
}

// DisconnectDB closes the MongoDB connection
func DisconnectDB() {
    if Client != nil {
        if err := Client.Disconnect(context.Background()); err != nil {
            log.Printf("Error disconnecting from MongoDB: %v", err)
        }
    }
}