package config

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	Address  string `json:"address"`
	Database struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		Database string `json:"database"`
	} `json:"database"`
	Twitch struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
	} `json:"twitch"`
	WebhookSecret string `json:"webhook_secret"`
	ExternalUrl   string `json:"external_url"`
}

var c *Config

func init() {
	jsonBytes, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatalf("failed to read config.json: %v", err)
	}
	c = &Config{}
	err = json.Unmarshal(jsonBytes, c)
	if err != nil {
		log.Fatalf("failed to unmarshal config.json: %v", err)
	}
}

func C() *Config {
	return c
}
