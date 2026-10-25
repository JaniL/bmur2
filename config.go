package bmur2

import (
	"fmt"
	"os"
)

type Config struct {
	BotToken      string
	WebhookSecret string
	LastfmAPIKey  string
	LastfmUser    string
}

func loadConfig() (*Config, error) {
	botToken, err := requiredEnv("BMUR_TOKEN")
	if err != nil {
		return nil, err
	}
	webhookSecret, err := requiredEnv("TELEGRAM_WEBHOOK_SECRET")
	if err != nil {
		return nil, err
	}
	lastfmAPIKey, err := requiredEnv("BMUR_LASTFM_APIKEY")
	if err != nil {
		return nil, err
	}
	lastfmUser, err := requiredEnv("BMUR_LASTFM_USERNAME")
	if err != nil {
		return nil, err
	}
	return &Config{
		BotToken:      botToken,
		WebhookSecret: webhookSecret,
		LastfmAPIKey:  lastfmAPIKey,
		LastfmUser:    lastfmUser,
	}, nil
}

func requiredEnv(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", fmt.Errorf("missing %s environment variable", name)
	}
	return value, nil
}
