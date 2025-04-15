package test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/francescorizzello94/senior-fullstack-engineer-takehome/internal/take-home/handler"
	"github.com/francescorizzello94/senior-fullstack-engineer-takehome/internal/take-home/model"
	"github.com/francescorizzello94/senior-fullstack-engineer-takehome/internal/take-home/service"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockIngestService struct {
	mock.Mock
}

func (m *MockIngestService) IngestFile(ctx context.Context, filePath string) error {
	args := m.Called(ctx, filePath)
	return args.Error(0)
}

func (m *MockIngestService) IngestSingle(ctx context.Context, data *model.WeatherData) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

type MockQueryService struct {
	mock.Mock
}

func (m *MockQueryService) GetByDate(ctx context.Context, date time.Time, opts ...*service.QueryOptions) ([]*model.WeatherData, error) {
	args := m.Called(ctx, date, opts)
	return args.Get(0).([]*model.WeatherData), args.Error(1)
}

func (m *MockQueryService) GetByDateRange(ctx context.Context, start, end time.Time, opts ...*service.QueryOptions) ([]*model.WeatherData, error) {
	args := m.Called(ctx, start, end, opts)
	return args.Get(0).([]*model.WeatherData), args.Error(1)
}

type MockWebSocketHub struct {
	mock.Mock
}

func (m *MockWebSocketHub) HandleConnection(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
}

func (m *MockWebSocketHub) Broadcast(data *model.WeatherData) {
	m.Called(data)
}

func (m *MockWebSocketHub) Run(ctx context.Context) {
	m.Called(ctx)
}

func (m *MockWebSocketHub) Shutdown() {
	m.Called()
}

type TestHandler struct {
	*handler.HTTPHandler
	IngestSvc *MockIngestService
	QuerySvc  *MockQueryService
	WSHub     *MockWebSocketHub
}

func setupTestHandler() *TestHandler {
	logger := zap.NewNop()
	ingestSvc := &MockIngestService{}
	querySvc := &MockQueryService{}
	wsHub := &MockWebSocketHub{}

	return &TestHandler{
		HTTPHandler: handler.NewHTTPHandler(
			ingestSvc,
			querySvc,
			wsHub,
			logger,
		),
		IngestSvc: ingestSvc,
		QuerySvc:  querySvc,
		WSHub:     wsHub,
	}
}

func TestHTTPHandler_GetWeatherByDate(t *testing.T) {
	th := setupTestHandler()
	router := mux.NewRouter()
	th.RegisterRoutes(router)

	testDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	expectedData := []*model.WeatherData{
		{
			Date:        testDate,
			Temperature: 22.5,
			Humidity:    75.5,
		},
	}

	t.Run("successful request", func(t *testing.T) {
		th.QuerySvc.On("GetByDate", mock.Anything, testDate, mock.Anything).Return(expectedData, nil)

		req := httptest.NewRequest("GET", "/api/v1/weather/2023-01-01", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response []*model.WeatherData
		assert.NoError(t, json.NewDecoder(w.Body).Decode(&response))
		assert.Len(t, response, 1)
		assert.Equal(t, expectedData[0], response[0])
		th.QuerySvc.AssertExpectations(t)
	})

	t.Run("invalid date format", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/weather/invalid-date", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHTTPHandler_GetWeatherByDateRange(t *testing.T) {
	th := setupTestHandler()
	router := mux.NewRouter()
	th.RegisterRoutes(router)

	startDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)
	expectedData := []*model.WeatherData{
		{
			Date:        startDate,
			Temperature: 22.5,
			Humidity:    75.5,
		},
		{
			Date:        endDate,
			Temperature: 23.5,
			Humidity:    76.5,
		},
	}

	t.Run("successful request", func(t *testing.T) {
		th.QuerySvc.On("GetByDateRange", mock.Anything, startDate, endDate, mock.Anything).Return(expectedData, nil)

		req := httptest.NewRequest("GET", "/api/v1/weather?from=2023-01-01&to=2023-01-02", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response []*model.WeatherData
		assert.NoError(t, json.NewDecoder(w.Body).Decode(&response))
		assert.Len(t, response, 2)
		th.QuerySvc.AssertExpectations(t)
	})

	t.Run("missing_date_parameters", func(t *testing.T) {
		// Create request with only the 'from' parameter
		req := httptest.NewRequest("GET", "/api/v1/weather?from=2023-01-01", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

	})
}

func TestHTTPHandler_IngestWeatherData(t *testing.T) {
	th := setupTestHandler()
	router := mux.NewRouter()
	th.RegisterRoutes(router)

	testData := &model.WeatherData{
		Date:        time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		Temperature: 22.5,
		Humidity:    75.5,
	}

	t.Run("successful ingestion", func(t *testing.T) {
		th.IngestSvc.On("IngestSingle", mock.Anything, testData).Return(nil)
		th.WSHub.On("Broadcast", testData).Return()

		body := `{"date":"2023-01-01T00:00:00Z","temperature":22.5,"humidity":75.5}`
		req := httptest.NewRequest("POST", "/api/v1/weather", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		th.IngestSvc.AssertExpectations(t)
		th.WSHub.AssertExpectations(t)
	})

	t.Run("invalid_content_type", func(t *testing.T) {
		// create request with invalid content type
		req := httptest.NewRequest("POST", "/api/v1/weather", strings.NewReader(`{"date":"2023-01-01","temperature":22.5,"humidity":75.5}`))
		req.Header.Set("Content-Type", "text/plain") // Set invalid content type
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

	})
}
