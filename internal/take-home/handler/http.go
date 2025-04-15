package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/francescorizzello94/senior-fullstack-engineer-takehome/internal/take-home/model"
	"github.com/francescorizzello94/senior-fullstack-engineer-takehome/internal/take-home/service"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type HTTPHandler struct {
	ingestSvc *service.IngestService
	querySvc  *service.QueryService
	wsHub     *WebSocketHub
	logger    *zap.Logger
}

func NewHTTPHandler(
	ingestSvc *service.IngestService,
	querySvc *service.QueryService,
	wsHub *WebSocketHub,
	logger *zap.Logger,
) *HTTPHandler {
	return &HTTPHandler{
		ingestSvc: ingestSvc,
		querySvc:  querySvc,
		wsHub:     wsHub,
		logger:    logger.Named("http_handler"),
	}
}

func (h *HTTPHandler) RegisterRoutes(router *mux.Router) {
	apiRouter := router.PathPrefix("/api/v1").Subrouter()

	// API endpoints
	apiRouter.HandleFunc("/weather", h.ingestWeatherData).
		Methods("POST").
		Headers("Content-Type", "application/json")

	apiRouter.HandleFunc("/weather/{date}", h.getWeatherByDate).
		Methods("GET")

	apiRouter.HandleFunc("/weather", h.getWeatherByDateRange).
		Methods("GET").
		Queries(
			"from", "{from:[0-9]{4}-[0-9]{2}-[0-9]{2}}",
			"to", "{to:[0-9]{4}-[0-9]{2}-[0-9]{2}}",
		)

	// WebSocket endpoint
	apiRouter.HandleFunc("/weather/ws", h.wsHub.HandleConnection)
}

func (h *HTTPHandler) ingestWeatherData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var data model.WeatherData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		h.logger.Warn("Invalid request payload", zap.Error(err))
		respondWithError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	if err := h.ingestSvc.IngestSingle(ctx, &data); err != nil {
		h.logger.Error("Ingestion failed", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to ingest data")
		return
	}

	// broadcast to WebSocket clients
	h.wsHub.Broadcast(&data)

	respondWithJSON(w, http.StatusCreated, data)
}

// create a new slice with only the requested fields included
func filterFields(data []*model.WeatherData, fields []string) []map[string]any {
	fieldMap := make(map[string]bool)
	for _, f := range fields {
		fieldMap[f] = true
	}

	// always include date field
	fieldMap["date"] = true

	result := make([]map[string]any, len(data))
	for i, item := range data {
		filtered := make(map[string]any)

		filtered["date"] = item.Date

		// only include requested fields
		if fieldMap["temperature"] {
			filtered["temperature"] = item.Temperature
		}
		if fieldMap["humidity"] {
			filtered["humidity"] = item.Humidity
		}
		// here more fields would appear as the data model grows
		// projection would be espcially useful for large documents with many fields

		result[i] = filtered
	}

	return result
}

func (h *HTTPHandler) getWeatherByDate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// parse and validate date
	date, err := time.Parse("2006-01-02", mux.Vars(r)["date"])
	if err != nil {
		h.logger.Warn("Invalid date format", zap.Error(err))
		respondWithError(w, http.StatusBadRequest, "Invalid date format (YYYY-MM-DD)")
		return
	}

	// build query options from request
	opts := buildQueryOptionsFromRequest(r)

	data, err := h.querySvc.GetByDate(ctx, date, opts)
	if err != nil {
		h.logger.Error("Query failed", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve data")
		return
	}

	if len(data) == 0 {
		respondWithError(w, http.StatusNotFound, "No data found for specified date")
		return
	}

	// check if filtering fields is needed
	fieldsParam := r.URL.Query().Get("fields")
	if fieldsParam != "" {
		fields := splitCommaSeparated(fieldsParam)
		filteredData := filterFields(data, fields)
		respondWithJSON(w, http.StatusOK, filteredData)
		return
	}

	respondWithJSON(w, http.StatusOK, data)
}

func (h *HTTPHandler) getWeatherByDateRange(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse and validate from and to dates from query parameters
	from, err := time.Parse("2006-01-02", r.URL.Query().Get("from"))
	if err != nil {
		h.logger.Warn("Invalid from date", zap.Error(err))
		respondWithError(w, http.StatusBadRequest, "Invalid 'from' date format")
		return
	}

	to, err := time.Parse("2006-01-02", r.URL.Query().Get("to"))
	if err != nil {
		h.logger.Warn("Invalid to date", zap.Error(err))
		respondWithError(w, http.StatusBadRequest, "Invalid 'to' date format")
		return
	}

	// build query options
	opts := buildQueryOptionsFromRequest(r)

	data, err := h.querySvc.GetByDateRange(ctx, from, to, opts)
	if err != nil {
		h.logger.Error("Query failed", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve data")
		return
	}

	if len(data) == 0 {
		respondWithError(w, http.StatusNotFound, "No data found for specified range")
		return
	}

	// check if filtering fields is needed
	fieldsParam := r.URL.Query().Get("fields")
	if fieldsParam != "" {
		fields := splitCommaSeparated(fieldsParam)
		filteredData := filterFields(data, fields)
		respondWithJSON(w, http.StatusOK, filteredData)
		return
	}

	respondWithJSON(w, http.StatusOK, data)
}

func buildQueryOptionsFromRequest(r *http.Request) *service.QueryOptions {
	opts := &service.QueryOptions{
		ExcludeID: true, // exclude ID by default
	}

	// field projection
	if fields := r.URL.Query().Get("fields"); fields != "" {
		opts.Fields = splitCommaSeparated(fields)
	}

	// pagination
	if page, err := strconv.ParseInt(r.URL.Query().Get("page"), 10, 64); err == nil && page > 0 {
		opts.Pagination.Page = page
	}

	if limit, err := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64); err == nil && limit > 0 {
		opts.Pagination.Limit = limit
	}

	return opts
}

// helper function to split a comma-separated string and trim spaces
// important for performance when documents grow large and many fields are requested

func splitCommaSeparated(s string) []string {
	if s == "" {
		return nil
	}

	parts := strings.Split(s, ",")
	for i, part := range parts {
		parts[i] = strings.TrimSpace(part)
	}

	return parts
}

func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}
