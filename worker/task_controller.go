package worker

import (
	"context"
	"sync"
	"time"

	"github.com/seniorcat/scraper/database"
	"github.com/seniorcat/scraper/entity"
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
	Category   *entity.Category
	RetryCount int
}

// Result представляет результат выполнения задачи
type Result struct {
	TaskID  string
	Recipes []entity.Recipe
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

	wg        sync.WaitGroup
	DBService database.DBServiceInterface // Используем интерфейс вместо структуры
}

// NewTaskController создает новый экземпляр TaskController
func NewTaskController(categoryWorker *CategoryWorker, workersCount int, logger *zap.Logger, retryInterval time.Duration, maxRetries int, dbService database.DBServiceInterface) *TaskController {
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
		DBService:     dbService,
	}
}

// InitWorkerPool инициализирует пул воркеров
func (tc *TaskController) InitWorkerPool(maxRecipes int, rps int, timeout time.Duration) {
	// Создаем воркеры и добавляем их в пул
	for i := 0; i < tc.WorkersCount; i++ {
		worker := NewRecipeWorker(tc.Logger, maxRecipes, rps, timeout)
		tc.RecipeWorkers = append(tc.RecipeWorkers, worker)

		// Запускаем каждого воркера в отдельной горутине
		go func(w *RecipeWorker) {
			defer tc.wg.Done() // После завершения работы воркера уменьшаем счетчик WaitGroup
			w.ProcessTasks(tc.TaskQueue, tc.ResultQueue)
		}(worker)
	}

	tc.Logger.Info("Worker pool initialized", zap.Int("workers_count", tc.WorkersCount))
}

// Start запускает контроллер задач для обработки всех задач из очереди
func (tc *TaskController) Start(maxRecipes int, rps int, timeout time.Duration) {
	tc.wg.Add(tc.WorkersCount)
	// Инициализация пула воркеров
	tc.InitWorkerPool(maxRecipes, rps, timeout)

	// Запуск обработки результатов
	go tc.ProcessResults()
}

// Stop завершает работу контроллера задач
func (tc *TaskController) Stop() {
	close(tc.TaskQueue)
	close(tc.ResultQueue)
	tc.wg.Wait()
}

// ProcessResults обрабатывает результаты из канала ResultQueue и сохраняет их в базу данных
func (tc *TaskController) ProcessResults() {
	ctx := context.Background()
	for result := range tc.ResultQueue {
		// Логирование результата
		tc.Logger.Info("Result received", zap.String("task_id", result.TaskID), zap.Int("recipes_count", len(result.Recipes)))

		// Сохранение рецептов в базу данных
		if err := tc.DBService.SaveRecipes(ctx, result.Recipes); err != nil {
			tc.Logger.Error("Failed to save recipes", zap.Error(err))
		} else {
			tc.Logger.Info("Recipes saved successfully", zap.String("task_id", result.TaskID))
		}
	}
}
