package cache

import (
	"sync"
)

// Кеш в памяти для хранения данных
type MemoryCache struct {
	mu    sync.RWMutex
	cache map[string]bool
}

// NewMemoryCache создает новый кеш в памяти
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		cache: make(map[string]bool),
	}
}

// Set добавляет элемент в кеш
func (m *MemoryCache) Set(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cache[key] = true
}

// Exists проверяет, существует ли элемент в кеше
func (m *MemoryCache) Exists(key string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cache[key]
}
