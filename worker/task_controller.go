package worker

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

// Status определяет состояние задачи или воркера
type Status string

const (
	StatusIdle       Status = "Idle"
	StatusBusy       Status = "Busy"
	StatusError      Status = "Error"
	StatusPending    Status = "Pending"
	StatusInProgress Status = "In Progress"
	StatusCompleted  Status = "Completed"
)

// Task представляет собой задачу, которая должна быть обработана воркером
type Task struct {
	ID         string
	Type       string
	Category   *Category
	RetryCount int
}

// Result представляет результат выполнения задачи
type Result struct {
	TaskID  string
	Recipes []Recipe
}

// TaskController управляет распределением задач между воркерами
type TaskController struct {
	CategoryWorker *CategoryWorker
	RecipeWorker   *RecipeWorker

	// Каналы для задач и результатов
	TaskQueue   chan Task
	ResultQueue chan Result

	Logger *zap.Logger
	wg     sync.WaitGroup

	retryInterval time.Duration
	maxRetries    int
}

// NewTaskController создает новый экземпляр TaskController
func NewTaskController(categoryWorker *CategoryWorker, recipeWorker *RecipeWorker, logger *zap.Logger, retryInterval time.Duration, maxRetries int) *TaskController {
	return &TaskController{
		CategoryWorker: categoryWorker,
		RecipeWorker:   recipeWorker,

		// Инициализация каналов
		TaskQueue:   make(chan Task, 100),
		ResultQueue: make(chan Result, 100),

		Logger:        logger,
		retryInterval: retryInterval,
		maxRetries:    maxRetries,
	}
}

// Start запускает контроллер задач для обработки всех задач из очереди
func (tc *TaskController) Start() {
	tc.Logger.Info("Task Controller started")
	tc.wg.Add(1)
	go tc.processTasks()
	go tc.ProcessResults() // Запускаем обработку результатов
	tc.wg.Wait()
}

// Stop завершает работу контроллера задач
func (tc *TaskController) Stop() {
	close(tc.TaskQueue)
	close(tc.ResultQueue)
	tc.wg.Done()
}

// processTasks распределяет задачи между воркерами
func (tc *TaskController) processTasks() {
	for task := range tc.TaskQueue {
		tc.Logger.Info("Task received", zap.String("task_type", task.Type))

		switch task.Type {
		case "category":
			tc.updateWorkerStatus("CategoryWorker", StatusBusy)
			err := tc.processCategoryTask()
			if err != nil {
				tc.handleTaskError(task)
			}
			tc.updateWorkerStatus("CategoryWorker", StatusIdle)
		case "recipe":
			tc.updateWorkerStatus("RecipeWorker", StatusBusy)
			if task.Category != nil {
				err := tc.processRecipeTask(*task.Category)
				if err != nil {
					tc.handleTaskError(task)
				}
			}
			tc.updateWorkerStatus("RecipeWorker", StatusIdle)
		}
	}
}

// ProcessResults обрабатывает результаты из канала ResultQueue
func (tc *TaskController) ProcessResults() {
	for result := range tc.ResultQueue {
		// Обрабатываем результат выполнения задачи
		tc.Logger.Info("Result received", zap.String("task_id", result.TaskID), zap.Int("recipes_count", len(result.Recipes)))

		// Дополнительная логика обработки результатов
	}
}

// handleTaskError обрабатывает ошибки задачи и планирует ее перезапуск
func (tc *TaskController) handleTaskError(task Task) {
	tc.Logger.Error("Task failed", zap.String("type", task.Type), zap.String("id", task.ID))

	if task.RetryCount < tc.maxRetries {
		task.RetryCount++
		go func() {
			time.Sleep(tc.retryInterval)
			tc.retryTask(task)
		}()
	} else {
		tc.Logger.Error("Max retry attempts reached for task", zap.String("id", task.ID))
		tc.updateTaskStatus(task, StatusError)
	}
}

// retryTask перезапускает задачу с обновленным статусом
func (tc *TaskController) retryTask(task Task) {
	tc.updateTaskStatus(task, StatusPending)
	tc.TaskQueue <- task
}

// updateTaskStatus обновляет статус задачи
func (tc *TaskController) updateTaskStatus(task Task, status Status) {
	// Логирование обновления статуса
	tc.Logger.Info("Task status updated", zap.String("task_id", task.ID), zap.String("status", string(status)))
}

// processCategoryTask обрабатывает задачи парсинга категорий
func (tc *TaskController) processCategoryTask() error {
	tc.Logger.Info("Processing category task")
	categories, err := tc.CategoryWorker.Start()
	if err != nil {
		return err
	}

	tc.Logger.Info("Successfully processed category task", zap.Int("count", len(categories)))

	for _, category := range categories {
		task := Task{
			ID:       category.Name,
			Type:     "recipe",
			Category: &category,
		}

		// Отправляем задачу в канал TaskQueue
		tc.TaskQueue <- task
	}
	return nil
}

// processRecipeTask обрабатывает задачи парсинга рецептов
func (tc *TaskController) processRecipeTask(category Category) error {
	tc.Logger.Info("Processing recipe task", zap.String("Category", category.Name))
	recipes, err := tc.RecipeWorker.Start(category)
	if err != nil {
		return err
	}

	// Отправляем результаты парсинга в канал результатов
	tc.ResultQueue <- Result{
		TaskID:  category.Name,
		Recipes: recipes,
	}

	tc.Logger.Info("Successfully processed recipe task", zap.Int("count", len(recipes)), zap.String("Category", category.Name))
	return nil
}

// updateWorkerStatus обновляет статус воркера
func (tc *TaskController) updateWorkerStatus(workerName string, status Status) {
	tc.Logger.Info("Worker status updated", zap.String("worker", workerName), zap.String("status", string(status)))
}
