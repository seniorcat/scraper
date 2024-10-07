package worker

import (
	"github.com/gocolly/colly" // Используем Colly для парсинга
	"go.uber.org/zap"          // Zap для логирования
)

// Category хранит информацию о категории
type Category struct {
	Name string
	Href string
}

// CategoryWorker отвечает за парсинг категорий
type CategoryWorker struct {
	Collector *colly.Collector
	Logger    *zap.Logger
}

// NewCategoryWorker создает новый экземпляр CategoryWorker
func NewCategoryWorker(logger *zap.Logger) *CategoryWorker {
	return &CategoryWorker{
		Collector: colly.NewCollector(), // Инициализация Colly Collector
		Logger:    logger,               // Инициализация логера
	}
}

// Start запускает парсинг категорий
func (w *CategoryWorker) Start() ([]Category, error) {
	var categories []Category

	// Парсинг данных о категориях
	w.Collector.OnHTML(".emotion-c3fqwx", func(e *colly.HTMLElement) {
		category := Category{
			Name: e.ChildText("a h3"),      // Извлечение названия категории
			Href: e.ChildAttr("a", "href"), // Извлечение ссылки
		}
		w.Logger.Info("Category found", zap.String("Name", category.Name))
		categories = append(categories, category) // Добавление категории в список
	})

	// URL для посещения и запуска парсинга категорий
	w.Collector.Visit("https://eda.ru")

	return categories, nil
}
