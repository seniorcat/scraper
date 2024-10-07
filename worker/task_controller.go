package worker

import (
	"sync"

	"go.uber.org/zap"
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
		switch task.Type {
		case "category":
			tc.processCategoryTask()
		case "recipe":
			if task.Category != nil {
				tc.processRecipeTask(*task.Category)
			}
		}
	}
}

// processCategoryTask обрабатывает задачи парсинга категорий
func (tc *TaskController) processCategoryTask() {
	tc.Logger.Info("Processing category task")
	categories, err := tc.CategoryWorker.Start()
	if err != nil {
		tc.Logger.Error("Failed to process category task", zap.Error(err))
		return
	}

	// Передаем задачи парсинга рецептов в очередь
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

	// Обработка результата
	for _, recipe := range recipes {
		tc.Logger.Info("Recipe processed", zap.String("Name", recipe.Name))
	}
}
