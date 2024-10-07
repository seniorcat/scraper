package worker_test

import (
	"testing"
	"time"

	"github.com/seniorcat/scraper/worker"
	"go.uber.org/zap"
)

func TestTaskProcessing(t *testing.T) {
	logger, _ := zap.NewProduction()
	categoryWorker := worker.NewCategoryWorker(logger, 5*time.Second)
	recipeWorker := worker.NewRecipeWorker(logger, 10, 5*time.Second)

	taskController := worker.NewTaskController(categoryWorker, recipeWorker, logger, 2*time.Second, 3)

	// Запуск обработчика задач
	go taskController.Start()

	// Добавление задачи
	taskController.TaskQueue <- worker.Task{ID: "test-category", Type: "category"}

	select {
	case result := <-taskController.ResultQueue:
		if result.TaskID != "test-category" {
			t.Errorf("Expected task ID 'test-category', got '%s'", result.TaskID)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Timeout waiting for result")
	}

	taskController.Stop()
}
