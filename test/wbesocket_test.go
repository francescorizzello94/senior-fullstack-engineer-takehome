package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/francescorizzello94/senior-fullstack-engineer-takehome/internal/take-home/handler"
	"github.com/francescorizzello94/senior-fullstack-engineer-takehome/internal/take-home/model"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestWebSocketHub_Run(t *testing.T) {
	logger := zap.NewNop()
	hub := handler.NewWebSocketHub(logger).(*handler.WebSocketHubImpl)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	// test that the hub shuts down properly
	time.Sleep(100 * time.Millisecond) // start
	cancel()                           // signal shutdown
	time.Sleep(100 * time.Millisecond) // shutdown

}

func TestWebSocketHub_Broadcast(t *testing.T) {
	logger := zap.NewNop()
	hub := handler.NewWebSocketHub(logger)

	data := &model.WeatherData{
		Date:        time.Now().UTC(),
		Temperature: 22.5,
		Humidity:    75.5,
	}

	// test that broadcast doesn't panic when there are no clients
	hub.Broadcast(data)

	// Successful test is one that doesn't panic
}

func TestWebSocketHub_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := zap.NewNop()
	hub := handler.NewWebSocketHub(logger)

	server := httptest.NewServer(http.HandlerFunc(hub.HandleConnection))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)

	// convert HTTP URL to ws URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	dialer := websocket.DefaultDialer
	ws, _, err := dialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer ws.Close()

	testData := &model.WeatherData{
		Date:        time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		Temperature: 22.5,
		Humidity:    75.5,
	}

	type wsResponse struct {
		data *model.WeatherData
		err  error
	}

	responseChan := make(chan wsResponse, 1)
	go func() {
		var receivedData model.WeatherData
		err := ws.ReadJSON(&receivedData)
		responseChan <- wsResponse{data: &receivedData, err: err}
	}()

	hub.Broadcast(testData)

	select {
	case resp := <-responseChan:
		assert.NoError(t, resp.err)
		assert.Equal(t, testData.Date.Format(time.RFC3339), resp.data.Date.Format(time.RFC3339))
		assert.Equal(t, testData.Temperature, resp.data.Temperature)
		assert.Equal(t, testData.Humidity, resp.data.Humidity)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for WebSocket response")
	}
}
