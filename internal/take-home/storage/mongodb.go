package storage

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type MongoDBRepository struct {
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
}

func Connect(ctx context.Context, uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to create MongoDB client: %w", err)
	}

	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	defer pingCancel()
	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return client, nil
}

func NewMongoDBRepository(client *mongo.Client) *MongoDBRepository {
	db := client.Database("oofone-se-take-home")
	return &MongoDBRepository{
		client:     client,
		database:   db,
		collection: db.Collection("weather_data"),
	}
}

func (r *MongoDBRepository) InsertWeatherData(ctx context.Context, data any) error {
	insertCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.collection.InsertOne(insertCtx, data)
	if err != nil {
		return fmt.Errorf("insert into collection '%s' failed: %w", r.collection.Name(), err)
	}
	return nil
}

func (r *MongoDBRepository) CloseConnection(ctx context.Context) error {
	disconnectCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return r.client.Disconnect(disconnectCtx)
}
