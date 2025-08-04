// Так как в кэше хранится история заказа, при изменении статуса заказа кэш автоматически инвалидируется,
// чтобы предотвратить использование устаревших данных
package order_cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type OrderCacheService struct {
	client *redis.Client
	ttl    time.Duration
}

func New(client *redis.Client, ttl time.Duration) *OrderCacheService {
	return &OrderCacheService{
		client: client,
		ttl:    ttl,
	}
}

func (s *OrderCacheService) Set(ctx context.Context, key string, value string) error {
	return s.client.Set(ctx, key, value, s.ttl).Err()
}

func (s *OrderCacheService) Get(ctx context.Context, key string) (string, error) {
	return s.client.Get(ctx, key).Result()
}

func (s *OrderCacheService) Delete(ctx context.Context, key string) error {
	return s.client.Del(ctx, key).Err()
}
