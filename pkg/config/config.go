package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Env         string
	Country     string
	ApiUrl      string
	ApiKey      string
	ProcessMode string
}

func LoadConfig(env *string) (Config, error) {

	err := godotenv.Load(*env)
	if err != nil {
		log.Printf("Error loading %s file", *env)
	}

	environment := os.Getenv("ENV")

	country := os.Getenv("COUNTRY")
	if country == "" {
		return Config{}, fmt.Errorf("COUNTRY cannot be empty")
	}

	apiUrl := os.Getenv("API_URL")
	if apiUrl == "" {
		return Config{}, fmt.Errorf("API_URL cannot be empty")
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		return Config{}, fmt.Errorf("API_KEY cannot be empty")
	}

	processMode := os.Getenv("PROCESS_MODE")

	return Config{
		Env:         environment,
		Country:     country,
		ApiUrl:      apiUrl,
		ApiKey:      apiKey,
		ProcessMode: processMode,
	}, nil
}
