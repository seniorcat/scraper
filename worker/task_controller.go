package worker

import (
	"sync"

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
	Type     string
	Category *Category
}

// TaskController управляет распределением задач между воркерами
type TaskController struct {
	CategoryWorker *CategoryWorker
	RecipeWorker   *RecipeWorker
	TaskQueue      chan Task
	Logger         *zap.Logger
	wg             sync.WaitGroup
}

// NewTaskController создает новый экземпляр TaskController
func NewTaskController(categoryWorker *CategoryWorker, recipeWorker *RecipeWorker, logger *zap.Logger) *TaskController {
	return &TaskController{
		CategoryWorker: categoryWorker,
		RecipeWorker:   recipeWorker,
		TaskQueue:      make(chan Task, 100), // Очередь задач
		Logger:         logger,
	}
}

// Start запускает контроллер задач для обработки всех задач из очереди
func (tc *TaskController) Start() {
	tc.Logger.Info("Task Controller started")
	tc.wg.Add(1)
	go tc.processTasks()
	tc.wg.Wait()
}

// Stop завершает работу контроллера задач
func (tc *TaskController) Stop() {
	close(tc.TaskQueue)
	tc.wg.Done()
}

// processTasks распределяет задачи между воркерами
func (tc *TaskController) processTasks() {
	for task := range tc.TaskQueue {
		tc.Logger.Info("Task received", zap.String("task_type", task.Type))

		switch task.Type {
		case "category":
			tc.updateWorkerStatus("CategoryWorker", StatusBusy)
			tc.processCategoryTask()
			tc.updateWorkerStatus("CategoryWorker", StatusIdle)
		case "recipe":
			tc.updateWorkerStatus("RecipeWorker", StatusBusy)
			if task.Category != nil {
				tc.processRecipeTask(*task.Category)
			}
			tc.updateWorkerStatus("RecipeWorker", StatusIdle)
		}
	}
}

// Обновление статуса воркера
func (tc *TaskController) updateWorkerStatus(workerName string, status Status) {
	tc.Logger.Info("Worker status updated", zap.String("worker", workerName), zap.String("status", string(status)))
}

// processCategoryTask обрабатывает задачи парсинга категорий
func (tc *TaskController) processCategoryTask() {
	tc.Logger.Info("Processing category task")
	categories, err := tc.CategoryWorker.Start()
	if err != nil {
		tc.Logger.Error("Failed to process category task", zap.Error(err))
		return
	}

	tc.Logger.Info("Successfully processed category task", zap.Int("count", len(categories)))

	for _, category := range categories {
		tc.TaskQueue <- Task{Type: "recipe", Category: &category}
	}
}

// processRecipeTask обрабатывает задачи парсинга рецептов
func (tc *TaskController) processRecipeTask(category Category) {
	tc.Logger.Info("Processing recipe task", zap.String("Category", category.Name))
	recipes, err := tc.RecipeWorker.Start(category)
	if err != nil {
		tc.Logger.Error("Failed to process recipe task", zap.String("category", category.Name), zap.Error(err))
		return
	}

	tc.Logger.Info("Successfully processed recipe task", zap.Int("count", len(recipes)), zap.String("Category", category.Name))
}
