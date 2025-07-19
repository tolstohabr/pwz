package redis

import (
	rds "github.com/redis/go-redis/v9"
)

func NewClient(cfg Config) *rds.Client {
	client := rds.NewClient(&rds.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return client
}
