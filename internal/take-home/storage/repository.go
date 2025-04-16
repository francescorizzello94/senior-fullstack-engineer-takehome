package storage

import (
	"context"
	"time"

	"github.com/francescorizzello94/senior-fullstack-engineer-takehome/internal/take-home/model"
)

// interface defines the behaviour in relation to the database

type WeatherRepository interface {
	InsertWeatherData(ctx context.Context, data any) error
	GetByDate(ctx context.Context, date time.Time, opts ...*QueryOptions) ([]*model.WeatherData, error)
	GetByDateRange(ctx context.Context, start, end time.Time, opts ...*QueryOptions) ([]*model.WeatherData, error)
	CloseConnection(ctx context.Context) error
}
