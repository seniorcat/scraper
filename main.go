package main

import (
	"fmt"

	"github.com/seniorcat/scraper/cmd"
)

func main() {
	cli := cmd.NewCLI()

	// Регистрация команды "run"
	cli.RegisterCommand("run", "Запуск парсера", func(args []string) {
		if err := cmd.RunParser(args); err != nil {
			// Обработка ошибки, если аргументы некорректные или другая проблема
			fmt.Printf("Ошибка при запуске парсера: %v\n", err)
		}
	})

	// Добавляем команду "help" для справки
	cli.RegisterCommand("help", "Вывод справки по командам", func(args []string) {
		cli.Run()
	})

	// Запуск CLI
	cli.Run()
}
