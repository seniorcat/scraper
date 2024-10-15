package cmd

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/seniorcat/scraper/config"
	"github.com/seniorcat/scraper/database"
	"github.com/seniorcat/scraper/entity"
	"github.com/seniorcat/scraper/pkg/cache"
	"github.com/seniorcat/scraper/worker"
	"go.uber.org/zap"
)

// RunParser запускает контроллер задач и управляет процессом парсинга
func RunParser() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Не удалось инициализировать логгер: %v", err)
	}
	defer logger.Sync()

	// Загрузка конфигурации
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		logger.Fatal("Ошибка загрузки конфигурации", zap.Error(err))
	}

	// Инициализация базы данных
	dbService, err := database.NewDBService(cfg.Database.URL)
	if err != nil {
		logger.Fatal("Ошибка подключения к базе данных", zap.Error(err))
	}

	// Считывание параметров из конфигурации
	timeout := cfg.Worker.Timeout
	maxRecipes := cfg.Worker.MaxRecipes
	retryInterval := cfg.Worker.RetryInterval
	maxRetries := cfg.Worker.MaxRetries
	concurrency := cfg.Worker.Concurrency
	rps := cfg.Worker.RPS

	// Создание кеша
	cache := cache.NewMemoryCache()

	// Создание воркера для категорий
	categoryWorker := worker.NewCategoryWorker(logger, rps, time.Duration(timeout)*time.Second, cache)

	// Создание контроллера задач с DI для работы с базой данных
	taskController := worker.NewTaskController(categoryWorker, concurrency, logger, time.Duration(retryInterval)*time.Second, maxRetries, dbService)

	// Запуск контроллера задач
	go taskController.Start(maxRecipes, rps, time.Duration(timeout)*time.Second)

	// Логирование запуска задачи
	logger.Info("Adding category parsing task to the queue")

	// Создаем канал для категорий
	categoryQueue := make(chan entity.Category)

	// Запускаем парсинг категорий в отдельной горутине и передаем категории в канал
	go func() {
		err := categoryWorker.Start(categoryQueue) // Передаем канал в Start
		if err != nil {
			logger.Fatal("Ошибка парсинга категорий", zap.Error(err))
		}
	}()

	// Обрабатываем категории: отправляем их на сохранение и добавляем задачи на парсинг рецептов
	go func() {
		for category := range categoryQueue {
			// Отправляем категорию на асинхронное сохранение
			dbService.CategorySaveChan <- []entity.Category{category}

			// Добавляем задачу на парсинг рецептов
			taskController.TaskQueue <- worker.Task{
				ID:       category.Name,
				Type:     "recipe",
				Category: &category,
			}
		}
	}()

	// Обработка сигналов для корректной остановки воркеров
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	// Ожидание сигнала остановки
	<-stopChan

	// Закрытие каналов для завершения работы воркеров
	close(dbService.CategorySaveChan)
	close(dbService.RecipeSaveChan)

	// Остановка контроллера задач
	taskController.Stop()
	logger.Info("Парсинг завершен.")
}
