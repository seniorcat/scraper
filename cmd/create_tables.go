package cmd

import (
	"context"
	"log"

	"github.com/seniorcat/scraper/config"
	"github.com/seniorcat/scraper/database"
	"go.uber.org/zap"
)

// CreateTables создает таблицы в базе данных
func CreateTables() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Не удалось инициализировать логгер: %v", err)
	}
	defer logger.Sync()

	// Загрузка конфигурации базы данных
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		logger.Fatal("Ошибка загрузки конфигурации", zap.Error(err))
	}

	// Подключение к базе данных
	dbService, err := database.NewDBService(cfg.Database.URL)
	if err != nil {
		logger.Fatal("Не удалось подключиться к базе данных", zap.Error(err))
	}

	// Создание таблиц
	err = dbService.CreateTables(context.Background())
	if err != nil {
		logger.Fatal("Не удалось создать таблицы", zap.Error(err))
	}

	logger.Info("Таблицы успешно созданы")
}
