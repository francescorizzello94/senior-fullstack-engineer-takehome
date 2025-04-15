package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/francescorizzello94/senior-fullstack-engineer-takehome/config"
	"github.com/francescorizzello94/senior-fullstack-engineer-takehome/internal/take-home/handler"
	"github.com/francescorizzello94/senior-fullstack-engineer-takehome/internal/take-home/service"
	"github.com/francescorizzello94/senior-fullstack-engineer-takehome/internal/take-home/storage"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// graceful shutdown context declaration
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// load config
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	// connect to db
	mongoClient, err := storage.Connect(ctx, cfg.MongoURI)
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB", zap.Error(err))
	}
	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			logger.Error("Error disconnecting from MongoDB", zap.Error(err))
		}
	}()

	// init repo
	repo := storage.NewMongoDBRepository(mongoClient)

	// init services
	ingestService := service.NewIngestService(*repo)
	queryService := service.NewQueryService(repo)

	// init WebSocket
	wsHub := handler.NewWebSocketHub(logger)

	// run websocket in separate goroutine
	go wsHub.Run(ctx)

	// init HTTP handler
	httpHandler := handler.NewHTTPHandler(
		ingestService,
		queryService,
		wsHub,
		logger,
	)

	// create router + register routes
	router := mux.NewRouter()
	httpHandler.RegisterRoutes(router)

	// init HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// start server in a goroutine
	go func() {
		logger.Info("Starting server", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	// load initial data from weather.dat
	go func() {
		logger.Info("Loading initial weather data")
		if err := ingestService.IngestFile(ctx, "data/weather.dat"); err != nil {
			logger.Error("Failed to ingest initial data", zap.Error(err))
		} else {
			logger.Info("Initial data loaded successfully")
		}
	}()

	// wait for termination signal
	<-ctx.Done()
	logger.Info("Shutdown signal received")

	// timeout context for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// attempt graceful shutdown
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown failed", zap.Error(err))
	}

	logger.Info("Server gracefully stopped")
}
