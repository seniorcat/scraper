package worker

import (
	"time"

	"github.com/gocolly/colly"
	"github.com/seniorcat/scraper/entity"
	"github.com/seniorcat/scraper/pkg/cache"
	"github.com/seniorcat/scraper/pkg/metrics"
	"go.uber.org/zap"
)

// CategoryParser отвечает за логику парсинга категорий
type CategoryParser struct {
	Collector *colly.Collector
	Logger    *zap.Logger
	Limiter   *RateLimiter
	timeout   time.Duration
	Cache     *cache.MemoryCache
}

// NewCategoryParser создает новый экземпляр CategoryParser
func NewCategoryParser(logger *zap.Logger, rps int, timeout time.Duration, cache *cache.MemoryCache) *CategoryParser {
	return &CategoryParser{
		Collector: colly.NewCollector(),
		Logger:    logger,
		Limiter:   NewRateLimiter(rps),
		timeout:   timeout,
		Cache:     cache,
	}
}

// ParseCategories выполняет сбор всех категорий
func (p *CategoryParser) ParseCategories() ([]entity.Category, error) {
	var categories []entity.Category
	p.Collector.OnHTML(".emotion-18mh8uc .emotion-c3fqwx", func(e *colly.HTMLElement) {
		// Увеличиваем счетчик запросов
		metrics.RequestCounter.Inc()
		// Извлечение имени категории, как было описано ранее
		categoryName := e.DOM.Find("a .emotion-1ooehk6").Clone().Children().Remove().End().Text()

		// Проверка через кеш, была ли категория уже обработана
		if p.Cache.Exists(categoryName) {
			p.Logger.Info("Category already cached, skipping", zap.String("Name", categoryName))
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

		// Добавление в кеш
		p.Cache.Set(categoryName)

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
func NewCategoryWorker(logger *zap.Logger, rps int, timeout time.Duration, cache *cache.MemoryCache) *CategoryWorker {
	parser := NewCategoryParser(logger, rps, timeout, cache)
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
