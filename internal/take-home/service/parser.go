package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/francescorizzello94/senior-fullstack-engineer-takehome/internal/take-home/model"
)

type WeatherParser struct{}

func NewWeatherParser() *WeatherParser {
	return &WeatherParser{}
}

func (p *WeatherParser) ParseStream(ctx context.Context, r io.Reader, handler func(data *model.WeatherData) error) error {
	scanner := bufio.NewScanner(r)
	lineNum := 0

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			lineNum++
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			data, err := p.parseLine(line)
			if err != nil {
				return fmt.Errorf("line %d: %w", lineNum, err)
			}

			if err := handler(data); err != nil {
				return fmt.Errorf("handler error line %d: %w", lineNum, err)
			}
		}
	}

	return scanner.Err()
}

func (p *WeatherParser) parseLine(line string) (*model.WeatherData, error) {
	parts := strings.Fields(line)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid format, expected 3 columns")
	}

	date, err := time.Parse("2002-02-05", parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid date: %w", err)
	}

	temp, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid temperature: %w", err)
	}

	humidity, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid humidity: %w", err)
	}

	data := &model.WeatherData{
		Date:        date,
		Temperature: temp,
		Humidity:    humidity,
	}

	if err := data.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return data, nil
}
