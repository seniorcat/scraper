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
	RecipeWorkers  []*RecipeWorker // Пул воркеров для обработки рецептов

	TaskQueue   chan Task
	ResultQueue chan Result

	WorkersCount int // Количество воркеров
	Logger       *zap.Logger

	retryInterval time.Duration
	maxRetries    int

	wg sync.WaitGroup
}

// NewTaskController создает новый экземпляр TaskController
func NewTaskController(categoryWorker *CategoryWorker, workersCount int, logger *zap.Logger, retryInterval time.Duration, maxRetries int) *TaskController {
	return &TaskController{
		CategoryWorker: categoryWorker,
		RecipeWorkers:  make([]*RecipeWorker, 0, workersCount), // Создаем слайс для пула воркеров

		// Инициализация каналов с ограничением на количество задач
		TaskQueue:   make(chan Task, 100),   // Лимит на 100 задач в очереди
		ResultQueue: make(chan Result, 100), // Лимит на 100 результатов в очереди

		WorkersCount:  workersCount,
		Logger:        logger,
		retryInterval: retryInterval,
		maxRetries:    maxRetries,
	}
}

// InitWorkerPool инициализирует пул воркеров
func (tc *TaskController) InitWorkerPool(maxRecipes int, timeout time.Duration) {
	// Создаем воркеры и добавляем их в пул
	for i := 0; i < tc.WorkersCount; i++ {
		worker := NewRecipeWorker(tc.Logger, maxRecipes, timeout)
		tc.RecipeWorkers = append(tc.RecipeWorkers, worker)

		// Запускаем каждого воркера в отдельной горутине
		go worker.ProcessTasks(tc.TaskQueue, tc.ResultQueue)
	}

	tc.Logger.Info("Worker pool initialized", zap.Int("workers_count", tc.WorkersCount))
}

// Start запускает контроллер задач для обработки всех задач из очереди
func (tc *TaskController) Start(maxRecipes int, timeout time.Duration) {
	// Инициализация пула воркеров
	tc.InitWorkerPool(maxRecipes, timeout)

	// Запуск обработки результатов
	go tc.ProcessResults()
}

// Stop завершает работу контроллера задач
func (tc *TaskController) Stop() {
	close(tc.TaskQueue)
	close(tc.ResultQueue)
	tc.wg.Done()
}

// ProcessResults обрабатывает результаты из канала ResultQueue
func (tc *TaskController) ProcessResults() {
	for result := range tc.ResultQueue {
		// Обрабатываем результат выполнения задачи
		tc.Logger.Info("Result received", zap.String("task_id", result.TaskID), zap.Int("recipes_count", len(result.Recipes)))

		// Дополнительная логика обработки результатов
	}
}
