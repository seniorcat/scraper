package worker

import (
	"testing"

	"github.com/seniorcat/scraper/pkg/cache"
	"go.uber.org/zap"
)

// TestCategoryWorkerStart проверяет корректность парсинга категорий
func TestCategoryWorkerStart(t *testing.T) {
	logger := zap.NewNop()             // Используем no-op логгер для тестов
	memCache := cache.NewMemoryCache() // Создаем новый кеш в памяти
	categoryWorker := NewCategoryWorker(logger, 5, 10, memCache)

	// Запуск парсинга категорий
	categories, err := categoryWorker.Start()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Проверка, что получили хотя бы одну категорию
	if len(categories) == 0 {
		t.Errorf("Expected at least one category, got %d", len(categories))
	}
}
