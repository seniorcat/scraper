package worker

import (
	"time"

	"github.com/gocolly/colly"
	"go.uber.org/zap"
)

// Category хранит информацию о категории
type Category struct {
	Name string
	Href string
}

// CategoryParser отвечает за логику парсинга категорий
type CategoryParser struct {
	Collector *colly.Collector
	Logger    *zap.Logger
	timeout   time.Duration
}

// NewCategoryParser создает новый экземпляр CategoryParser
func NewCategoryParser(logger *zap.Logger, timeout time.Duration) *CategoryParser {
	return &CategoryParser{
		Collector: colly.NewCollector(),
		Logger:    logger,
		timeout:   timeout,
	}
}

// ParseCategories выполняет сбор всех категорий
func (p *CategoryParser) ParseCategories() ([]Category, error) {
	var categories []Category

	p.Collector.OnHTML(".emotion-c3fqwx", func(e *colly.HTMLElement) {
		category := Category{
			Name: e.ChildText("a h3"),
			Href: e.ChildAttr("a", "href"),
		}
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
func NewCategoryWorker(logger *zap.Logger, timeout time.Duration) *CategoryWorker {
	parser := NewCategoryParser(logger, timeout)
	return &CategoryWorker{Parser: parser}
}

// Start запускает воркер для парсинга категорий
func (w *CategoryWorker) Start() ([]Category, error) {
	categories, err := w.Parser.ParseCategories()
	if err != nil {
		w.Parser.Logger.Error("Failed to parse categories", zap.Error(err))
		return nil, err
	}
	return categories, nil
}
