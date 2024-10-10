package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/seniorcat/scraper/config"
	"github.com/seniorcat/scraper/database"
	"github.com/seniorcat/scraper/worker"
	"go.uber.org/zap"
)

// RunParser запускает контроллер задач и управляет процессом парсинга
func RunParser() {
	// Инициализация логгера zap
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

	ctx := context.Background()

	// Считывание параметров из конфигурации
	timeout := cfg.Worker.Timeout
	maxRecipes := cfg.Worker.MaxRecipes
	retryInterval := cfg.Worker.RetryInterval
	maxRetries := cfg.Worker.MaxRetries
	concurrency := cfg.Worker.Concurrency
	rps := cfg.Worker.RPS

	// Создание воркера для категорий с передачей лимитера в воркеры
	categoryWorker := worker.NewCategoryWorker(logger, rps, time.Duration(timeout)*time.Second)

	// Создание контроллера задач с DI для работы с базой данных
	taskController := worker.NewTaskController(categoryWorker, concurrency, logger, time.Duration(retryInterval)*time.Second, maxRetries, dbService)

	// Запуск контроллера задач и пула воркеров
	go taskController.Start(maxRecipes, rps, time.Duration(timeout)*time.Second)

	// Логирование запуска задачи
	logger.Info("Adding category parsing task to the queue")

	// Парсинг категорий
	categories, err := categoryWorker.Start()
	if err != nil {
		logger.Fatal("Ошибка парсинга категорий", zap.Error(err))
	}

	// Сохранение категорий в базе данных
	if err := dbService.SaveCategories(ctx, categories); err != nil {
		logger.Fatal("Ошибка сохранения категорий", zap.Error(err))
	}

	// Добавление задач для парсинга рецептов в TaskQueue
	for _, category := range categories {
		logger.Info("Adding recipe parsing task for category", zap.String("category_name", category.Name))
		taskController.TaskQueue <- worker.Task{
			ID:       category.Name,
			Type:     "recipe",
			Category: &category,
		}
	}

	// Обработка сигналов для корректной остановки воркеров
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	// Ожидание сигнала остановки
	<-stopChan

	// Остановка контроллера задач
	taskController.Stop()
	logger.Info("Парсинг завершен.")
}
