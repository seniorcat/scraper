package worker_test

import (
	"context"
	"testing"
	"time"

	"github.com/seniorcat/scraper/entity"
	"github.com/seniorcat/scraper/pkg/cache"
	"github.com/seniorcat/scraper/worker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockDBService - это моковая реализация DBServiceInterface для тестов
type MockDBService struct {
	mock.Mock
}

// SaveCategories - моковая реализация
func (m *MockDBService) SaveCategories(ctx context.Context, categories []entity.Category) error {
	args := m.Called(ctx, categories)
	return args.Error(0)
}

// SaveRecipes - моковая реализация
func (m *MockDBService) SaveRecipes(ctx context.Context, recipes []entity.Recipe) error {
	args := m.Called(ctx, recipes)
	return args.Error(0)
}

func TestTaskController_ProcessResults(t *testing.T) {
	// Инициализация мока базы данных
	mockDB := new(MockDBService)

	// Инициализация логгера
	logger, _ := zap.NewDevelopment()

	// Настройка поведения мока: рецепты будут успешно сохранены
	mockDB.On("SaveRecipes", mock.Anything, mock.Anything).Return(nil)

	// Создание контроллера задач
	tc := worker.NewTaskController(nil, 2, logger, time.Second, 3, mockDB)

	// Запуск обработки результатов в отдельной горутине
	go tc.ProcessResults()

	// Отправка результата в очередь для обработки
	recipes := []entity.Recipe{
		{Name: "Recipe1", Href: "/recepty/zavtraki/draniki-iz-batata-187448"},
		{Name: "Recipe2", Href: "/recepty/zavtraki/grechnevij-zavtrak-22397"},
	}

	result := worker.Result{
		TaskID:  "task1",
		Recipes: recipes,
	}

	tc.ResultQueue <- result
	close(tc.ResultQueue)

	// Ожидание обработки результата
	time.Sleep(100 * time.Millisecond)

	// Проверка, что SaveRecipes был вызван
	mockDB.AssertCalled(t, "SaveRecipes", mock.Anything, recipes)
}

func TestTaskController_AddTaskAndProcess(t *testing.T) {
	// Инициализация мока базы данных
	mockDB := new(MockDBService)

	// Инициализация логгера
	logger, _ := zap.NewDevelopment()

	// Настройка поведения мока: рецепты будут успешно сохранены
	mockDB.On("SaveRecipes", mock.Anything, mock.Anything).Return(nil)
	memCache := cache.NewMemoryCache() // Создаем новый кеш в памяти

	// Создание воркера категории и контроллера задач
	categoryWorker := worker.NewCategoryWorker(logger, 10, time.Second, memCache)
	tc := worker.NewTaskController(categoryWorker, 2, logger, time.Second, 3, mockDB)

	// Запуск контроллера задач
	go tc.Start(10, 5, time.Second)

	// Добавление задачи в очередь
	task := worker.Task{
		ID:       "category1",
		Type:     "recipe",
		Category: &entity.Category{Name: "Category1", Href: "/recepty/zavtraki"},
	}
	tc.TaskQueue <- task

	// Создание воркера для обработки задачи
	recipeWorker := worker.NewRecipeWorker(logger, 10, 5, time.Second)
	tc.RecipeWorkers = append(tc.RecipeWorkers, recipeWorker)

	// Обработка задачи воркером
	recipes := []entity.Recipe{
		{Name: "Recipe1", Href: "/recepty/zavtraki/draniki-iz-batata-187448"},
		{Name: "Recipe2", Href: "/recepty/zavtraki/grechnevij-zavtrak-22397"},
	}

	// Отправляем результат в ResultQueue для обработки
	tc.ResultQueue <- worker.Result{
		TaskID:  task.ID,
		Recipes: recipes,
	}

	// Закрытие очереди задач и результатов
	close(tc.TaskQueue)
	close(tc.ResultQueue)

	// Ожидание обработки результата
	time.Sleep(100 * time.Millisecond)

	// Проверка, что SaveRecipes был вызван
	mockDB.AssertCalled(t, "SaveRecipes", mock.Anything, recipes)
	mockDB.AssertExpectations(t)
}

func TestTaskController_Stop(t *testing.T) {
	// Инициализация мока базы данных
	mockDB := new(MockDBService)
	memCache := cache.NewMemoryCache() // Создаем новый кеш в памяти

	// Инициализация логгера
	logger, _ := zap.NewDevelopment()

	// Создание воркера категории и контроллера задач
	categoryWorker := worker.NewCategoryWorker(logger, 10, time.Second, memCache)
	tc := worker.NewTaskController(categoryWorker, 2, logger, time.Second, 3, mockDB)

	// Запуск контроллера задач
	go tc.Start(10, 5, time.Second)

	// Остановка контроллера задач
	tc.Stop()

	// Проверка, что очереди закрыты
	_, ok := <-tc.TaskQueue
	assert.False(t, ok)

	_, ok = <-tc.ResultQueue
	assert.False(t, ok)
}

func TestRecipeWorkerProcessTasks(t *testing.T) {
	logger := zap.NewNop() // Используем no-op логгер для тестов
	recipeWorker := worker.NewRecipeWorker(logger, 5, 10, time.Second*10)

	taskQueue := make(chan worker.Task, 1)
	resultQueue := make(chan worker.Result, 1)

	// Пример категории для теста
	category := entity.Category{
		Name: "Test Category",
		Href: "/recepty/zavtraki",
	}

	// Задача для обработки
	taskQueue <- worker.Task{
		ID:       "1",
		Type:     "recipe",
		Category: &category,
	}
	close(taskQueue)

	// Запуск обработки задач
	go recipeWorker.ProcessTasks(taskQueue, resultQueue)

	result := <-resultQueue

	// Проверка, что результат содержит рецепты
	if len(result.Recipes) == 0 {
		t.Errorf("Expected at least one recipe, got %d", len(result.Recipes))
	}

	// Проверка корректного ID задачи в результате
	if result.TaskID != "1" {
		t.Errorf("Expected task ID '1', got %s", result.TaskID)
	}
}
