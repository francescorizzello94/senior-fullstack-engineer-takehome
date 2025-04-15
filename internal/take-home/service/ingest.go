package service

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/francescorizzello94/senior-fullstack-engineer-takehome/internal/take-home/model"
	"github.com/francescorizzello94/senior-fullstack-engineer-takehome/internal/take-home/storage"
)

type IngestServiceInterface interface {
	IngestFile(ctx context.Context, filePath string) error
	IngestSingle(ctx context.Context, data *model.WeatherData) error
}

type QueryServiceInterface interface {
	GetByDate(ctx context.Context, date time.Time, opts ...*QueryOptions) ([]*model.WeatherData, error)
	GetByDateRange(ctx context.Context, start, end time.Time, opts ...*QueryOptions) ([]*model.WeatherData, error)
}

type IngestService struct {
	repo   storage.WeatherRepository
	parser *WeatherParser
}

func NewIngestService(repo storage.WeatherRepository) IngestServiceInterface {
	return &IngestService{
		repo:   repo,
		parser: NewWeatherParser(),
	}
}

func (s *IngestService) IngestFile(ctx context.Context, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return s.parser.ParseStream(ctx, file, func(data *model.WeatherData) error {
		if err := s.repo.InsertWeatherData(ctx, data); err != nil {
			return fmt.Errorf("failed to insert data: %w", err)
		}
		return nil
	})
}

func (s *IngestService) IngestSingle(ctx context.Context, data *model.WeatherData) error {
	if err := data.Validate(); err != nil {
		return fmt.Errorf("invalid data: %w", err)
	}

	data.Date = time.Date(
		data.Date.Year(),
		data.Date.Month(),
		data.Date.Day(),
		0, 0, 0, 0,
		time.UTC,
	)
	return s.repo.InsertWeatherData(ctx, data)
}
