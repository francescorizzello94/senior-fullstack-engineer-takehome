package storage

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"

	"github.com/francescorizzello94/senior-fullstack-engineer-takehome/internal/take-home/model"
)

// Index metadata constants
const (
	dateIndexName = "date_1"
)

type MongoDBRepository struct {
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
}

func Connect(ctx context.Context, uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().
		ApplyURI(uri).
		SetRetryWrites(true).
		SetMaxPoolSize(100).
		//SetSocketTimeout(30 * time.Second).
		SetServerSelectionTimeout(10 * time.Second))
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
	col := db.Collection("weather_data")

	// ensure indexes with existence check to improve performance in case of large datasets
	// avoids unnecessary index creation if it already exists
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ensureIndexes(ctx, col)

	return &MongoDBRepository{
		client:     client,
		database:   db,
		collection: col,
	}
}

func ensureIndexes(ctx context.Context, col *mongo.Collection) {
	// ensure indexes and avoid duplicates
	indexView := col.Indexes()
	cursor, err := indexView.List(ctx)
	if err != nil {
		fmt.Printf("Failed to list indexes: %v\n", err)
		return
	}
	defer cursor.Close(ctx)

	var indexExists bool
	for cursor.Next(ctx) {
		var index bson.M
		if err := cursor.Decode(&index); err != nil {
			fmt.Printf("Failed to decode index: %v\n", err)
			continue
		}

		if name, ok := index["name"].(string); ok && name == dateIndexName {
			indexExists = true
			break
		}
	}

	// create index if it doesn't exist and set unique constraint
	if !indexExists {
		_, err := indexView.CreateOne(ctx, mongo.IndexModel{
			Keys: bson.D{{Key: "date", Value: 1}},
			Options: options.Index().
				SetName(dateIndexName).
				SetUnique(true),
		})

		if err != nil {
			fmt.Printf("Failed to create index: %v\n", err)
		}
	}
}

// QueryOptions to provide control over projection and pagination
// projection capability added to future-proof for data model expansion
type QueryOptions struct {
	Projection bson.M
	Skip       *int64
	Limit      *int64
	Sort       bson.D
}

// return safe defaults
func DefaultQueryOptions() *QueryOptions {
	return &QueryOptions{
		Projection: bson.M{"_id": 0}, // Exclude ID by default, non-informative
		Sort:       bson.D{{Key: "date", Value: 1}},
	}
}

func (r *MongoDBRepository) GetByDate(
	ctx context.Context,
	date time.Time,
	opts ...*QueryOptions,
) ([]*model.WeatherData, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(24 * time.Hour)

	filter := bson.M{
		"date": bson.M{"$gte": start, "$lt": end}, // precise date aggregation, preferrable over $eq for performance
	}

	return r.queryWeatherData(ctx, filter, opts...)
}

func (r *MongoDBRepository) GetByDateRange(
	ctx context.Context,
	start, end time.Time,
	opts ...*QueryOptions,
) ([]*model.WeatherData, error) {
	filter := bson.M{
		"date": bson.M{"$gte": start, "$lte": end}, // precise date range aggregation, $lte instead of $lt as above because we are working with a range
	}

	return r.queryWeatherData(ctx, filter, opts...)
}

func (r *MongoDBRepository) queryWeatherData(
	ctx context.Context,
	filter bson.M,
	opts ...*QueryOptions,
) ([]*model.WeatherData, error) {
	findCtx, cancel := context.WithTimeout(ctx, 15*time.Second) // increased timeout for potentially large datasets
	defer cancel()

	// apply options or fallback to default
	queryOpts := DefaultQueryOptions()
	if len(opts) > 0 && opts[0] != nil {
		queryOpts = opts[0]
	}

	findOptions := options.Find().
		SetSort(queryOpts.Sort).
		SetBatchSize(1000)

	if queryOpts.Projection != nil {
		findOptions.SetProjection(queryOpts.Projection)
	}
	if queryOpts.Skip != nil {
		findOptions.SetSkip(*queryOpts.Skip)
	}
	if queryOpts.Limit != nil {
		findOptions.SetLimit(*queryOpts.Limit)
	}

	cursor, err := r.collection.Find(findCtx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("find operation failed: %w", err)
	}
	defer cursor.Close(ctx)

	var results []*model.WeatherData
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode results: %w", err)
	}

	return results, nil
}

func (r *MongoDBRepository) InsertWeatherData(ctx context.Context, data any) error {
	insertCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// mormalize date to midnight UTC for consistency
	if weatherData, ok := data.(*model.WeatherData); ok {
		weatherData.Date = time.Date(
			weatherData.Date.Year(),
			weatherData.Date.Month(),
			weatherData.Date.Day(),
			0, 0, 0, 0,
			time.UTC,
		)
		filter := bson.M{"date": weatherData.Date}
		update := bson.M{"$set": weatherData}
		opts := options.UpdateOne().SetUpsert(true)
		if _, err := r.collection.UpdateOne(insertCtx, filter, update, opts); err != nil {
			return fmt.Errorf("failed to upsert data into collection '%s': %w", r.collection.Name(), err)
		}
		return nil
	}
	return fmt.Errorf("invalid data type, expected *model.WeatherData")
}

func (r *MongoDBRepository) CloseConnection(ctx context.Context) error {
	disconnectCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return r.client.Disconnect(disconnectCtx)
}
