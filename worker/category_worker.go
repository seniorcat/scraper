package worker

import (
	"time"

	"github.com/gocolly/colly"
	"github.com/seniorcat/scraper/entity"
	"go.uber.org/zap"
)

// CategoryParser отвечает за логику парсинга категорий
type CategoryParser struct {
	Collector *colly.Collector
	Logger    *zap.Logger
	Limiter   *RateLimiter
	timeout   time.Duration
}

// NewCategoryParser создает новый экземпляр CategoryParser
func NewCategoryParser(logger *zap.Logger, rps int, timeout time.Duration) *CategoryParser {
	return &CategoryParser{
		Collector: colly.NewCollector(),
		Logger:    logger,
		Limiter:   NewRateLimiter(rps),
		timeout:   timeout,
	}
}

// ParseCategories выполняет сбор всех категорий
func (p *CategoryParser) ParseCategories() ([]entity.Category, error) {
	var categories []entity.Category
	// Используем карту для хранения уникальных категорий
	uniqueCategories := make(map[string]struct{})
	p.Collector.OnHTML(".emotion-18mh8uc .emotion-c3fqwx", func(e *colly.HTMLElement) {

		// Извлечение имени категории, как было описано ранее
		categoryName := e.DOM.Find("a .emotion-1ooehk6").Clone().Children().Remove().End().Text()

		// Проверка, существует ли такая категория уже
		if _, exists := uniqueCategories[categoryName]; exists {
			p.Logger.Info("Duplicate category found, skipping", zap.String("Name", categoryName))
			return
		}

		category := entity.Category{
			Name: categoryName,
			Href: e.ChildAttr("a", "href"),
		}

		// Нормализация данных категории
		category.Normalize()

		// Валидация категории
		if err := category.Validate(); err != nil {
			p.Logger.Error("Invalid category data", zap.Error(err))
			return
		}

		// Добавление категории в карту уникальных категорий
		uniqueCategories[categoryName] = struct{}{}

		p.Logger.Info("Category found", zap.String("Name", category.Name))
		categories = append(categories, category)
	})

	// URL для парсинга
	err := p.Collector.Visit("https://eda.ru")
	if err != nil {
		return nil, err
	}

	return categories, nil
}

// CategoryWorker управляет парсингом категорий
type CategoryWorker struct {
	Parser *CategoryParser
}

// NewCategoryWorker создает новый экземпляр CategoryWorker
func NewCategoryWorker(logger *zap.Logger, rps int, timeout time.Duration) *CategoryWorker {
	parser := NewCategoryParser(logger, rps, timeout)
	return &CategoryWorker{Parser: parser}
}

// Start запускает воркер для парсинга категорий
func (w *CategoryWorker) Start() ([]entity.Category, error) {
	categories, err := w.Parser.ParseCategories()
	if err != nil {
		w.Parser.Logger.Error("Failed to parse categories", zap.Error(err))
		return nil, err
	}
	return categories, nil
}
