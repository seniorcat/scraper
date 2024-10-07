package main

import (
	"flag"
	"fmt"

	"github.com/seniorcat/scraper/cmd"
)

func main() {
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
