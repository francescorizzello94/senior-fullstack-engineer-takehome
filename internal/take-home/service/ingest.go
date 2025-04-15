package service

import (
	"context"
	"fmt"
	"os"

	"github.com/francescorizzello94/senior-fullstack-engineer-takehome/internal/take-home/model"
	"github.com/francescorizzello94/senior-fullstack-engineer-takehome/internal/take-home/storage"
)

type IngestService struct {
	repo   storage.WeatherRepository
	parser *WeatherParser
}

func NewIngestService(repo storage.WeatherRepository) *IngestService {
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
	return s.repo.InsertWeatherData(ctx, data)
}
