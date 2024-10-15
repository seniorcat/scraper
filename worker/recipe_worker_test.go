package worker

import (
	"testing"
	"time"

	"github.com/seniorcat/scraper/entity"
	"go.uber.org/zap"
)

// TestRecipeWorkerStart проверяет корректность парсинга рецептов
func TestRecipeWorkerStart(t *testing.T) {
	logger := zap.NewNop()                                         // Используем no-op логгер для тестов
	recipeWorker := NewRecipeWorker(logger, 5, 10, time.Second*10) // Создаем новый RecipeWorker

	// Пример категории для парсинга рецептов
	category := entity.Category{
		Name: "Example Category",
		Href: "/recepty/zavtraki",
	}

	// Запуск парсинга рецептов из категории
	recipes, err := recipeWorker.Parser.ParseRecipes(category)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Проверка, что получили хотя бы один рецепт
	if len(recipes) == 0 {
		t.Errorf("Expected at least one recipe, got %d", len(recipes))
	}

	// Проверка, что количество рецептов не превышает максимального лимита
	if len(recipes) > recipeWorker.Parser.maxRecipes {
		t.Errorf("Expected no more than %d recipes, got %d", recipeWorker.Parser.maxRecipes, len(recipes))
	}
}
