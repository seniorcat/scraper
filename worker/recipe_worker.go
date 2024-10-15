package worker

import (
	"sync"
	"time"

	"github.com/gocolly/colly"
	"github.com/seniorcat/scraper/entity"
	"go.uber.org/zap"
)

// RecipeParser отвечает за логику парсинга рецептов
type RecipeParser struct {
	Collector  *colly.Collector
	Logger     *zap.Logger
	Limiter    *RateLimiter
	maxRecipes int
	timeout    time.Duration
}

// NewRecipeParser создает новый экземпляр RecipeParser
func NewRecipeParser(logger *zap.Logger, maxRecipes int, rps int, timeout time.Duration) *RecipeParser {
	return &RecipeParser{
		Collector:  colly.NewCollector(),
		Logger:     logger,
		Limiter:    NewRateLimiter(rps),
		maxRecipes: maxRecipes,
		timeout:    timeout,
	}
}

// ParseRecipes парсит рецепты для заданной категории
func (p *RecipeParser) ParseRecipes(category entity.Category) ([]entity.Recipe, error) {
	var recipes []entity.Recipe

	p.Limiter.TakeToken() // Ограничение скорости запросов

	p.Collector.OnHTML(".emotion-13pp0tv", func(e *colly.HTMLElement) {
		if len(recipes) >= p.maxRecipes {
			return // Прерывание парсинга, если достигнут лимит рецептов
		}
		recipe := entity.Recipe{
			Name: e.ChildAttr("img", "alt"),
			Href: e.Attr("href"),
		}

		// Нормализация данных рецепта
		recipe.Normalize()

		// Валидация рецепта
		if err := recipe.Validate(); err != nil {
			p.Logger.Error("Invalid recipe data", zap.Error(err))
			return
		}
		p.Logger.Info("Recipe found", zap.String("Name", recipe.Name))
		recipes = append(recipes, recipe)
	})

	// URL для парсинга
	err := p.Collector.Visit("https://eda.ru" + category.Href)
	if err != nil {
		return nil, err
	}

	return recipes, nil
}

// RecipeWorker управляет парсингом рецептов с синхронизацией
type RecipeWorker struct {
	Parser         *RecipeParser
	ProcessedCount int
	Mutex          *sync.Mutex // Добавляем мьютекс для синхронизации
}

// NewRecipeWorker создает новый экземпляр RecipeWorker
func NewRecipeWorker(logger *zap.Logger, maxRecipes int, rps int, timeout time.Duration) *RecipeWorker {
	parser := NewRecipeParser(logger, maxRecipes, rps, timeout)
	return &RecipeWorker{
		Parser: parser,
		Mutex:  &sync.Mutex{},
	}
}

// ProcessTasks запускает воркер для обработки задач и защищает доступ к счетчику
func (w *RecipeWorker) ProcessTasks(taskQueue chan Task, resultQueue chan Result) {
	for task := range taskQueue {
		if task.Type == "recipe" && task.Category != nil {
			recipes, err := w.Parser.ParseRecipes(*task.Category)
			if err != nil {
				w.Parser.Logger.Error("Failed to parse recipes", zap.String("category", task.Category.Name), zap.Error(err))
				continue
			}

			// Безопасное обновление счетчика обработанных рецептов
			w.Mutex.Lock()
			w.ProcessedCount += len(recipes)
			w.Mutex.Unlock()

			resultQueue <- Result{
				TaskID:  task.ID,
				Recipes: recipes,
			}
		}
	}
}
