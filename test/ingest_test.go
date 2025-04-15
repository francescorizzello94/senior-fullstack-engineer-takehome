package test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/francescorizzello94/senior-fullstack-engineer-takehome/internal/take-home/model"
	"github.com/francescorizzello94/senior-fullstack-engineer-takehome/internal/take-home/service"
	"github.com/francescorizzello94/senior-fullstack-engineer-takehome/internal/take-home/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDBRepository struct {
	mock.Mock
}

// mocks the repository's insert method
func (m *MockDBRepository) InsertWeatherData(ctx context.Context, data any) error {
	return m.Called(ctx, data).Error(0)
}

func (m *MockDBRepository) GetByDate(ctx context.Context, date time.Time, opts ...*storage.QueryOptions) ([]*model.WeatherData, error) {
	args := m.Called(ctx, date, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.WeatherData), args.Error(1)
}

func (m *MockDBRepository) GetByDateRange(ctx context.Context, start, end time.Time, opts ...*storage.QueryOptions) ([]*model.WeatherData, error) {
	args := m.Called(ctx, start, end, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.WeatherData), args.Error(1)
}

func (m *MockDBRepository) CloseConnection(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func TestIngestService(t *testing.T) {
	validData := &model.WeatherData{
		Date:        time.Now(),
		Temperature: 22.5,
		Humidity:    75.5,
	}

	t.Run("Valid data inserts successfully", func(t *testing.T) {
		repo := new(MockDBRepository)
		repo.On("InsertWeatherData", mock.Anything, validData).Return(nil)

		svc := service.NewIngestService(repo)
		err := svc.IngestSingle(context.Background(), validData)

		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("Invalid data fails validation", func(t *testing.T) {
		repo := new(MockDBRepository)
		// no need to set up expectations as validation should fail before repo is called

		svc := service.NewIngestService(repo)
		invalidData := *validData
		invalidData.Temperature = 150 // out of range

		err := svc.IngestSingle(context.Background(), &invalidData)
		assert.Error(t, err)
		// ensure repo was never called
		repo.AssertNotCalled(t, "InsertWeatherData")
	})

	t.Run("File ingestion parses lines correctly", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "weather*.dat")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		testData := "2023-01-01\t22.5\t75.5\n2023-01-02\t23.5\t76.5\n"
		if _, err := tmpFile.WriteString(testData); err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
		tmpFile.Close()

		repo := new(MockDBRepository)
		// set up expectations - should be called for each line in the file
		repo.On("InsertWeatherData", mock.Anything, mock.Anything).Return(nil).Times(2)

		svc := service.NewIngestService(repo)

		err = svc.IngestFile(context.Background(), tmpFile.Name())
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("File not found returns error", func(t *testing.T) {
		repo := new(MockDBRepository)
		svc := service.NewIngestService(repo)

		// use non-existent file path
		nonExistentPath := filepath.Join(os.TempDir(), "non_existent_file.dat")

		err := svc.IngestFile(context.Background(), nonExistentPath)
		assert.Error(t, err)
		// no repo calls expected
		repo.AssertNotCalled(t, "InsertWeatherData")
	})

	t.Run("Date is normalized to midnight UTC", func(t *testing.T) {
		repo := new(MockDBRepository)
		// Use custom matcher to verify date normalization
		repo.On("InsertWeatherData", mock.Anything, mock.MatchedBy(func(data any) bool {
			if wd, ok := data.(*model.WeatherData); ok {
				return wd.Date.Hour() == 0 && wd.Date.Minute() == 0 &&
					wd.Date.Second() == 0 && wd.Date.Nanosecond() == 0
			}
			return false
		})).Return(nil)

		svc := service.NewIngestService(repo)

		dataWithTime := &model.WeatherData{
			Date:        time.Date(2023, 10, 15, 14, 30, 45, 123000000, time.Local),
			Temperature: 22.5,
			Humidity:    75.5,
		}

		err := svc.IngestSingle(context.Background(), dataWithTime)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("Repository error is propagated", func(t *testing.T) {
		repo := new(MockDBRepository)
		expectedErr := assert.AnError // testify's built-in error
		repo.On("InsertWeatherData", mock.Anything, mock.Anything).Return(expectedErr)

		svc := service.NewIngestService(repo)

		err := svc.IngestSingle(context.Background(), validData)
		assert.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		repo.AssertExpectations(t)
	})
}

func BenchmarkIngestService_IngestSingle(b *testing.B) {
	data := &model.WeatherData{
		Date:        time.Now(),
		Temperature: 22.5,
		Humidity:    75.5,
	}

	repo := new(MockDBRepository)
	repo.On("InsertWeatherData", mock.Anything, mock.Anything).Return(nil)

	svc := service.NewIngestService(repo)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := svc.IngestSingle(context.Background(), data)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkIngestService_IngestFile(b *testing.B) {
	tmpFile, err := os.CreateTemp("", "weather_bench*.dat")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	for i := 0; i < 100; i++ {
		day := i%28 + 1
		month := (i % 12) + 1
		year := 2023
		temp := 20.0 + float64(i%10)
		humidity := 60.0 + float64(i%20)

		line := fmt.Sprintf("%04d-%02d-%02d\t%.2f\t%.2f\n", year, month, day, temp, humidity)
		if _, err := tmpFile.WriteString(line); err != nil {
			b.Fatalf("Failed to write to temp file: %v", err)
		}
	}
	tmpFile.Close()

	repo := new(MockDBRepository)
	repo.On("InsertWeatherData", mock.Anything, mock.Anything).Return(nil)

	svc := service.NewIngestService(repo)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := svc.IngestFile(context.Background(), tmpFile.Name())
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}
