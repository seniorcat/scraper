package cmd

import (
	"flag"
	"log"

	"github.com/seniorcat/scraper/worker"
	"go.uber.org/zap"
)

// parse запускает процесс парсинга в зависимости от параметров.
func parse(parseType, maxRecipes, concurrency int) {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Не удалось инициализировать логгер: %v", err)
	}
	defer logger.Sync() // Синхронизация логов перед завершением работы

	// Создание воркеров для категорий и рецептов
	categoryWorker := worker.NewCategoryWorker(logger)
	recipeWorker := worker.NewRecipeWorker(logger)

	switch parseType {
	case 1:
		// Полный парсинг: категории + рецепты
		logger.Info("Запуск полного парсинга...")

		// Парсинг категорий
		categories, err := categoryWorker.Start()
		if err != nil {
			logger.Error("Ошибка при парсинге категорий", zap.Error(err))
			return
		}

		// Логирование найденных категорий и парсинг рецептов
		for _, category := range categories {
			logger.Info("Категория", zap.String("Name", category.Name), zap.String("Href", category.Href))

			// Парсинг рецептов в каждой категории
			recipes, err := recipeWorker.Start(category)
			if err != nil {
				logger.Error("Ошибка при парсинге рецептов", zap.String("Category", category.Name), zap.Error(err))
				continue
			}

			// Логирование найденных рецептов
			for _, recipe := range recipes {
				logger.Info("Рецепт", zap.String("Name", recipe.Name), zap.String("Href", recipe.Href))
			}
		}
	case 2:
		// Парсинг только категорий
		logger.Info("Запуск парсинга категорий...")

		categories, err := categoryWorker.Start()
		if err != nil {
			logger.Error("Ошибка при парсинге категорий", zap.Error(err))
			return
		}

		// Логирование найденных категорий
		for _, category := range categories {
			logger.Info("Категория", zap.String("Name", category.Name), zap.String("Href", category.Href))
		}
	default:
		logger.Error("Неверный тип парсинга. Используйте 1 для полного парсинга или 2 для парсинга только категорий.")
	}
}

// RunParser запускает парсер с переданными аргументами.
func RunParser(args []string) error {
	var (
		parseType   int // Тип парсинга (1 - полный, 2 - только категории)
		maxRecipes  int // Количество рецептов для парсинга в каждой категории
		concurrency int // Количество одновременных потоков (горутин)
	)

	fs := flag.NewFlagSet("run", flag.ExitOnError)
	fs.IntVar(&parseType, "type", 1, "Тип парсинга: 1 - полный, 2 - только категории")
	fs.IntVar(&maxRecipes, "recipes", 10, "Количество рецептов для каждой категории")
	fs.IntVar(&concurrency, "concurrency", 5, "Количество одновременных потоков")

	if err := fs.Parse(args); err != nil {
		return err
	}

	log.Println("Запуск парсера с параметрами:")
	log.Printf("Тип парсинга: %d\n", parseType)
	log.Printf("Количество рецептов: %d\n", maxRecipes)
	log.Printf("Количество одновременных потоков: %d\n", concurrency)

	parse(parseType, maxRecipes, concurrency)

	return nil
}
