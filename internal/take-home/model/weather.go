package model

import (
	"fmt"
	"time"
)

type WeatherData struct {
	Date        time.Time `bson:"date" json:"date"`
	Temperature float64   `bson:"temperature" json:"temperature"`
	Humidity    float64   `bson:"humidity" json:"humidity"`
}

// validate weather data input

func (w *WeatherData) Validate() error {
	if w.Date.IsZero() {
		return fmt.Errorf("date is required")
	}
	if w.Temperature < -100 || w.Temperature > 100 {
		return fmt.Errorf("temperature outside valid range (-100 to 100)")
	}
	if w.Humidity < 0 || w.Humidity > 100 {
		return fmt.Errorf("humidity outside valid range (0 to 100)")
	}
	return nil
}
