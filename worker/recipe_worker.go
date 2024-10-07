package worker

import (
	"time"

	"github.com/gocolly/colly"
	"go.uber.org/zap"
)

// Recipe хранит информацию о рецепте
type Recipe struct {
	Name string
	Href string
}

// RecipeParser отвечает за логику парсинга рецептов
type RecipeParser struct {
	Collector  *colly.Collector
	Logger     *zap.Logger
	maxRecipes int
	timeout    time.Duration
}

// NewRecipeParser создает новый экземпляр RecipeParser
func NewRecipeParser(logger *zap.Logger, maxRecipes int, timeout time.Duration) *RecipeParser {
	return &RecipeParser{
		Collector:  colly.NewCollector(),
		Logger:     logger,
		maxRecipes: maxRecipes,
		timeout:    timeout,
	}
}

// ParseRecipes парсит рецепты для заданной категории
func (p *RecipeParser) ParseRecipes(category Category) ([]Recipe, error) {
	var recipes []Recipe

	p.Collector.OnHTML(".emotion-1j5xcrd", func(e *colly.HTMLElement) {
		if len(recipes) >= p.maxRecipes {
			return // Прерывание парсинга, если достигнут лимит рецептов
		}
		recipe := Recipe{
			Name: e.ChildText("a span"),
			Href: e.ChildAttr("a", "href"),
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

// RecipeWorker управляет парсингом рецептов
type RecipeWorker struct {
	Parser *RecipeParser
}

// NewRecipeWorker создает новый экземпляр RecipeWorker
func NewRecipeWorker(logger *zap.Logger, maxRecipes int, timeout time.Duration) *RecipeWorker {
	parser := NewRecipeParser(logger, maxRecipes, timeout)
	return &RecipeWorker{Parser: parser}
}

// Start запускает воркер для парсинга рецептов
func (w *RecipeWorker) Start(category Category) ([]Recipe, error) {
	recipes, err := w.Parser.ParseRecipes(category)
	if err != nil {
		w.Parser.Logger.Error("Failed to parse recipes", zap.String("category", category.Name), zap.Error(err))
		return nil, err
	}
	return recipes, nil
}
