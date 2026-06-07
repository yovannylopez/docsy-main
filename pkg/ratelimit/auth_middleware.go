package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"

	"github.com/yovannylopez/docsy-main/internal/shared/infrastructure/config"
	"github.com/yovannylopez/docsy-main/pkg/http_status"
	"github.com/yovannylopez/docsy-main/pkg/responses"
)

type allowBackend interface {
	allow(ctx context.Context, key string) (bool, error)
}

type redisBackend struct {
	client *redis.Client
	window time.Duration
	max    int
}

func (r *redisBackend) allow(ctx context.Context, key string) (bool, error) {
	k := "ratelimit:auth:" + key
	n, err := r.client.Incr(ctx, k).Result()
	if err != nil {
		return false, fmt.Errorf("redis incr %s: %w", k, err)
	}
	if n == 1 {
		_ = r.client.Expire(ctx, k, r.window)
	}
	return n <= int64(r.max), nil
}

func (r *redisBackend) close() error {
	return r.client.Close()
}

type memoryBackend struct {
	mu      sync.Mutex
	buckets map[string][]time.Time
	window  time.Duration
	max     int
}

func (m *memoryBackend) allow(_ context.Context, key string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-m.window)
	prev := m.buckets[key]
	kept := prev[:0]
	for _, ts := range prev {
		if ts.After(cutoff) {
			kept = append(kept, ts)
		}
	}
	if len(kept) >= m.max {
		m.buckets[key] = kept
		return false, nil
	}
	kept = append(kept, now)
	m.buckets[key] = kept
	return true, nil
}

// AuthRateLimiter expone el middleware de rate limiting y permite liberar recursos al cerrar.
type AuthRateLimiter struct {
	Middleware echo.MiddlewareFunc
	backend    allowBackend
}

// Close releases the backend resources (closes the Redis connection if configured).
func (a *AuthRateLimiter) Close() error {
	if rb, ok := a.backend.(*redisBackend); ok {
		return rb.close()
	}
	return nil
}

// NewAuthMiddleware crea un rate limiter por IP (Redis si está configurado; en memoria como fallback).
//
// Deprecated: prefer NewAuthRateLimiter which exposes Close for proper resource cleanup.
func NewAuthMiddleware(cfg config.RedisConfig) echo.MiddlewareFunc {
	return NewAuthRateLimiter(cfg).Middleware
}

// NewAuthRateLimiter creates an AuthRateLimiter with support for orderly resource cleanup.
func NewAuthRateLimiter(cfg config.RedisConfig) *AuthRateLimiter {
	window := time.Duration(cfg.AuthWindowSecs) * time.Second
	if window <= 0 {
		window = time.Minute
	}
	max := cfg.AuthMaxRequests
	if max <= 0 {
		max = 60
	}

	var be allowBackend
	if cfg.URL != "" {
		opt, err := redis.ParseURL(cfg.URL)
		if err == nil {
			be = &redisBackend{
				client: redis.NewClient(opt),
				window: window,
				max:    max,
			}
		}
	}
	if be == nil {
		be = &memoryBackend{
			buckets: make(map[string][]time.Time),
			window:  window,
			max:     max,
		}
	}

	mw := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := c.RealIP()
			ok, err := be.allow(c.Request().Context(), key)
			if err != nil {
				return next(c)
			}
			if !ok {
				return responses.EchoError(c, &http_status.TooManyRequests, "Too many attempts; please try again later")
			}
			return next(c)
		}
	}

	return &AuthRateLimiter{
		Middleware: mw,
		backend:    be,
	}
}
