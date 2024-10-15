package worker

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/seniorcat/scraper/pkg/cache"
	"go.uber.org/zap"
)

// TestWorkerIntegration проверяет корректность работы воркеров вместе
func TestWorkerIntegration(t *testing.T) {
	logger := zap.NewNop()             // Используем no-op логгер для тестов
	memCache := cache.NewMemoryCache() // Создаем новый кеш в памяти

	taskQueue := make(chan Task, 10)
	resultQueue := make(chan Result, 10)
	errChan := make(chan error, 1) // Канал для передачи ошибок из горутин

	// Инициализируем воркеры и контроллер задач
	categoryWorker := NewCategoryWorker(logger, 5, time.Second*10, memCache)
	recipeWorker := NewRecipeWorker(logger, 5, 10, time.Second*10)

	var wg sync.WaitGroup

	// Запуск воркеров
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Шаг 1: Парсинг категорий
		categories, err := categoryWorker.Start()
		if err != nil {
			errChan <- err // Передаем ошибку через канал
			return
		}

		// Проверка, что есть хотя бы одна категория
		if len(categories) == 0 {
			errChan <- fmt.Errorf("expected at least one category, got %d", len(categories))
			return
		}

		// Шаг 2: Добавление задач для парсинга рецептов в очередь задач
		for _, category := range categories {
			taskQueue <- Task{
				ID:       category.Name,
				Type:     "recipe",
				Category: &category,
			}
		}
		close(taskQueue) // Закрываем TaskQueue, когда все задачи отправлены
	}()

	wg.Add(1) // Увеличиваем счетчик для воркера рецептов
	go func() {
		defer wg.Done()

		recipeWorker.ProcessTasks(taskQueue, resultQueue)
	}()

	// Ждем завершения всех задач и горутин
	go func() {
		wg.Wait()          // Ждем завершения всех воркеров
		close(resultQueue) // Закрываем resultQueue только после завершения всех воркеров
	}()

	// Проверка, произошли ли ошибки в горутине
	select {
	case err := <-errChan:
		t.Fatalf("Test failed: %v", err)
	default:
		// Если ошибок не было, продолжаем проверку результатов
		for result := range resultQueue {
			if len(result.Recipes) == 0 {
				t.Errorf("Expected at least one recipe, got %d", len(result.Recipes))
			}
		}
	}
}
