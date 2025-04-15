package service

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/francescorizzello94/senior-fullstack-engineer-takehome/internal/take-home/model"
	"github.com/francescorizzello94/senior-fullstack-engineer-takehome/internal/take-home/storage"
)

type QueryService struct {
	repo storage.WeatherRepository
}

func NewQueryService(repo storage.WeatherRepository) *QueryService {
	return &QueryService{repo: repo}
}

type QueryOptions struct {
	Fields     []string // fields to be included/excluded
	ExcludeID  bool     // exclude ID unless needed
	Pagination struct {
		Page  int64
		Limit int64
	}
}

func (s *QueryService) GetByDate(
	ctx context.Context,
	date time.Time,
	opts ...*QueryOptions,
) ([]*model.WeatherData, error) {
	if date.IsZero() {
		return nil, fmt.Errorf("date cannot be zero")
	}

	// convert service-level options to storage-level options
	mongoOpts := buildMongoQueryOptions(opts...)

	data, err := s.repo.GetByDate(ctx, date, mongoOpts)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	return data, nil
}

func (s *QueryService) GetByDateRange(
	ctx context.Context,
	start, end time.Time,
	opts ...*QueryOptions,
) ([]*model.WeatherData, error) {
	switch {
	case start.IsZero() || end.IsZero():
		return nil, fmt.Errorf("both dates for the date range must be specified")
	case end.Before(start):
		return nil, fmt.Errorf("end date cannot be set prior to start date")
	case end.Sub(start) > 365*24*time.Hour:
		return nil, fmt.Errorf("date range may not exceed one year")
	}

	mongoOpts := buildMongoQueryOptions(opts...)

	data, err := s.repo.GetByDateRange(ctx, start, end, mongoOpts)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	return data, nil
}

func buildMongoQueryOptions(opts ...*QueryOptions) *storage.QueryOptions {
	if len(opts) == 0 || opts[0] == nil {
		return storage.DefaultQueryOptions()
	}

	serviceOpts := opts[0]
	mongoOpts := &storage.QueryOptions{
		Sort: bson.D{{Key: "date", Value: 1}},
	}

	// handle field projection conservatively
	// only include fields explicitly requested and if no fields are specified, default to nil (no projection)
	if len(serviceOpts.Fields) > 0 {
		projection := bson.M{}
		for _, field := range serviceOpts.Fields {
			projection[field] = 1
		}
		// always include the "date" field (and set _id exclusion explicitly)
		projection["date"] = 1
		projection["_id"] = 0

		mongoOpts.Projection = projection
	}

	// pagination
	if serviceOpts.Pagination.Limit > 0 {
		mongoOpts.Limit = &serviceOpts.Pagination.Limit
		skip := (serviceOpts.Pagination.Page - 1) * serviceOpts.Pagination.Limit
		mongoOpts.Skip = &skip
	}

	return mongoOpts
}
