package worker

import (
	"github.com/gocolly/colly" // Используем Colly для парсинга
	"go.uber.org/zap"          // Zap для логирования
)

// Recipe хранит информацию о рецепте
type Recipe struct {
	Name string
	Href string
}

// RecipeWorker отвечает за парсинг рецептов
type RecipeWorker struct {
	Collector *colly.Collector
	Logger    *zap.Logger
}

// NewRecipeWorker создает новый экземпляр RecipeWorker
func NewRecipeWorker(logger *zap.Logger) *RecipeWorker {
	return &RecipeWorker{
		Collector: colly.NewCollector(), // Инициализация Colly Collector
		Logger:    logger,               // Инициализация логера
	}
}

// Start запускает парсинг рецептов из конкретной категории
func (w *RecipeWorker) Start(category Category) ([]Recipe, error) {
	var recipes []Recipe

	// Парсинг данных о рецептах
	w.Collector.OnHTML(".emotion-1j5xcrd", func(e *colly.HTMLElement) {
		recipe := Recipe{
			Name: e.ChildText("a span"),    // Извлечение названия рецепта
			Href: e.ChildAttr("a", "href"), // Извлечение ссылки на рецепт
		}
		w.Logger.Info("Recipe found", zap.String("Name", recipe.Name))
		recipes = append(recipes, recipe) // Добавление рецепта в список
	})

	// URL для посещения и запуска парсинга рецептов
	w.Collector.Visit("https://eda.ru" + category.Href)

	return recipes, nil
}
