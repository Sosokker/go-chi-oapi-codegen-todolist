package cache

import (
	"context"
	"log/slog"
	"time"

	"github.com/Sosokker/todolist-backend/internal/config"
	gocache "github.com/patrickmn/go-cache"
)

// Cache defines the interface for a caching layer
type Cache interface {
	Get(ctx context.Context, key string) (interface{}, bool)
	Set(ctx context.Context, key string, value interface{}, duration time.Duration)
	Delete(ctx context.Context, key string)
}

// memoryCache is an in-memory implementation of the Cache interface
type memoryCache struct {
	client            *gocache.Cache
	logger            *slog.Logger
	defaultExpiration time.Duration
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache(cfg config.CacheConfig, logger *slog.Logger) Cache {
	defaultExp := cfg.DefaultExpiration
	if defaultExp <= 0 {
		defaultExp = 5 * time.Minute
	}
	cleanupInterval := cfg.CleanupInterval
	if cleanupInterval <= 0 {
		cleanupInterval = 10 * time.Minute
	}

	c := gocache.New(defaultExp, cleanupInterval)
	logger.Info("In-memory cache initialized",
		"defaultExpiration", defaultExp,
		"cleanupInterval", cleanupInterval)
	return &memoryCache{
		client:            c,
		logger:            logger.With("component", "memoryCache"),
		defaultExpiration: defaultExp,
	}
}

func (m *memoryCache) Get(ctx context.Context, key string) (interface{}, bool) {
	val, found := m.client.Get(key)
	if found {
		m.logger.DebugContext(ctx, "Cache hit", "key", key)
	} else {
		m.logger.DebugContext(ctx, "Cache miss", "key", key)
	}
	return val, found
}

func (m *memoryCache) Set(ctx context.Context, key string, value interface{}, duration time.Duration) {
	exp := duration
	if exp <= 0 {
		exp = m.defaultExpiration
	}
	m.logger.DebugContext(ctx, "Setting cache", "key", key, "duration", exp)
	m.client.Set(key, value, exp)
}

func (m *memoryCache) Delete(ctx context.Context, key string) {
	m.logger.DebugContext(ctx, "Deleting cache", "key", key)
	m.client.Delete(key)
}
