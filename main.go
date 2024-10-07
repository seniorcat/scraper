package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/seniorcat/scraper/cmd"
	"github.com/seniorcat/scraper/config"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Использование конфигурации
	fmt.Printf("Database URL: %s\n", cfg.Database.URL)
	fmt.Printf("Worker Type: %d\n", cfg.Worker.Type)
	fmt.Printf("Max Recipes: %d\n", cfg.Worker.MaxRecipes)
	fmt.Printf("Timeout: %d\n", cfg.Worker.Timeout)
	fmt.Printf("Max Retries: %d\n", cfg.Worker.MaxRetries)
	fmt.Printf("Retry Interval: %d\n", cfg.Worker.RetryInterval)
	cli := cmd.NewCLI()

	// Команда "run" для запуска парсера
	cli.RegisterCommand("run", "Запуск парсера", func(args []string) {
		var configPath string
		fs := flag.NewFlagSet("run", flag.ExitOnError)
		fs.StringVar(&configPath, "config", "config/config.yaml", "Путь к файлу конфигурации")
		fs.Parse(args)
		fmt.Printf("Запуск парсера с конфигурацией: %s\n", configPath)
		// Здесь будет логика запуска парсера, например, загрузка конфигурации
	})

	// Команда "help" для отображения списка команд
	cli.RegisterCommand("help", "Вывод справки по командам", func(args []string) {
		cli.PrintHelp()
	})

	cli.Run()
}
