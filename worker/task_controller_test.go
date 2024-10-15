package worker_test

import (
	"context"
	"testing"
	"time"

	"github.com/seniorcat/scraper/entity"
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
		{Name: "Recipe1", Href: "https://example.com/recipe1"},
		{Name: "Recipe2", Href: "https://example.com/recipe2"},
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

	// Создание воркера категории и контроллера задач
	categoryWorker := worker.NewCategoryWorker(logger, 10, time.Second)
	tc := worker.NewTaskController(categoryWorker, 2, logger, time.Second, 3, mockDB)

	// Запуск контроллера задач
	go tc.Start(10, 5, time.Second)

	// Добавление задачи в очередь
	task := worker.Task{
		ID:       "category1",
		Type:     "recipe",
		Category: &entity.Category{Name: "Category1", Href: "https://example.com/category1"},
	}
	tc.TaskQueue <- task

	// Создание воркера для обработки задачи
	recipeWorker := worker.NewRecipeWorker(logger, 10, 5, time.Second)
	tc.RecipeWorkers = append(tc.RecipeWorkers, recipeWorker)

	// Обработка задачи воркером
	recipes := []entity.Recipe{
		{Name: "Recipe1", Href: "https://example.com/recipe1"},
		{Name: "Recipe2", Href: "https://example.com/recipe2"},
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

	// Инициализация логгера
	logger, _ := zap.NewDevelopment()

	// Создание воркера категории и контроллера задач
	categoryWorker := worker.NewCategoryWorker(logger, 10, time.Second)
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
