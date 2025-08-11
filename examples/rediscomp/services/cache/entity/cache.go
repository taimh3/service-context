package entity

import "time"

// CacheItem represents a cache item
type CacheItem struct {
	Key       string        `json:"key"`
	Value     string        `json:"value"`
	TTL       time.Duration `json:"ttl,omitempty"`
	ExpiresAt *time.Time    `json:"expires_at,omitempty"`
}

// SetCacheRequest represents the request to set a cache item
type SetCacheRequest struct {
	Key   string        `json:"key" binding:"required"`
	Value string        `json:"value" binding:"required"`
	TTL   time.Duration `json:"ttl,omitempty"` // TTL in seconds, 0 means no expiration
}

// CacheFilter represents filter for cache operations
type CacheFilter struct {
	Pattern string `json:"pattern,omitempty"` // Redis key pattern for searching
}
