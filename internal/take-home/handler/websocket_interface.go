package handler

import (
	"context"
	"net/http"

	"github.com/francescorizzello94/senior-fullstack-engineer-takehome/internal/take-home/model"
)

type WebSocketHub interface {
	Run(ctx context.Context)

	HandleConnection(w http.ResponseWriter, r *http.Request)

	Broadcast(data *model.WeatherData)
}
