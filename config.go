package main

import (
	"errors"
	"os"
)

type Config struct {
	Bucket    string
	AccessKey string
	SecretKey string
}

func loadConfig() (Config, error) {
	cfg := Config{
		Bucket:    os.Getenv("BUCKET"),
		AccessKey: os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
	}
	if cfg.Bucket == "" || cfg.AccessKey == "" || cfg.SecretKey == "" {
		return Config{}, errors.New("не указаны переменные окружения")
	}
	return cfg, nil
}
