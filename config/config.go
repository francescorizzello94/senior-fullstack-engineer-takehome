package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type ColumnDefinition struct {
	Description string `yaml:"description"`
	Unit        string `yaml:"unit"`
}

type Config struct {
	Port     string
	MongoURI string
	Columns  map[string]ColumnDefinition `yaml:"columns"`
}

func LoadConfig() (*Config, error) {

	if err := godotenv.Load(); err != nil {
		fmt.Println("Note: .env file not found, using environment variables")
	}
	// Load .env properties
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		return nil, fmt.Errorf("MONGO_URI must be set")
	}

	// Load YAML column definitions
	data, err := os.ReadFile("config/columns.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read columns.yaml: %w", err)
	}

	var yamlConfig struct {
		Columns map[string]ColumnDefinition `yaml:"columns"`
	}

	if err := yaml.Unmarshal(data, &yamlConfig); err != nil {
		return nil, fmt.Errorf("failed to parse columns.yaml: %w", err)
	}

	required := []string{"Date", "Temperature", "Humidity"}
	for _, col := range required {
		if _, exists := yamlConfig.Columns[col]; !exists {
			return nil, fmt.Errorf("missing required column in YAML: %s", col)
		}
	}

	return &Config{
		Port:     port,
		MongoURI: mongoURI,
		Columns:  yamlConfig.Columns,
	}, nil
}
