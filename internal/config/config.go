package config

import "time"

type Config struct {
	Port          int
	Cache         CacheConfig
	HackerNewsAPI HackerNewsAPIConfig
}

type CacheConfig struct {
	ItemTTL time.Duration
}

type HackerNewsAPIConfig struct {
	BaseURL      string
	ItemsPerPage int
	WorkerCount  int
}

func New() *Config {
	return &Config{
		Port: 3000,
		Cache: CacheConfig{
			ItemTTL: 2 * time.Minute,
		},
		HackerNewsAPI: HackerNewsAPIConfig{
			BaseURL:      "https://hacker-news.firebaseio.com/v0",
			ItemsPerPage: 30,
			WorkerCount:  10,
		},
	}
}
