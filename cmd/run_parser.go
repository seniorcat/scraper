package cmd

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/seniorcat/scraper/config"
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

	// Считывание параметров из конфигурации
	timeout := cfg.Worker.Timeout
	maxRecipes := cfg.Worker.MaxRecipes

	// Создание воркеров с использованием конфигурации
	categoryWorker := worker.NewCategoryWorker(logger, time.Duration(timeout)*time.Second)
	recipeWorker := worker.NewRecipeWorker(logger, int(maxRecipes), time.Duration(timeout)*time.Second)

	// Создание контроллера задач
	taskController := worker.NewTaskController(categoryWorker, recipeWorker, logger)

	// Запуск контроллера задач
	go taskController.Start()

	// Добавление задачи для парсинга категорий
	taskController.TaskQueue <- worker.Task{Type: "category"}

	// Обработка сигналов для корректной остановки воркеров
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	// Ожидание сигнала остановки
	<-stopChan

	// Остановка контроллера задач
	taskController.Stop()
	logger.Info("Парсинг завершен.")
}
