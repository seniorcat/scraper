package main

import (
	"github.com/seniorcat/scraper/cmd"
)

func main() {
	cli := cmd.NewCLI()

	// Регистрация команды "run"
	cli.RegisterCommand("run", "Запуск парсера", func(args []string) {
		cmd.RunParser()
	})

	// Добавляем команду "help" для справки
	cli.RegisterCommand("help", "Вывод справки по командам", func(args []string) {
		cli.Run()
	})

	// Запуск CLI
	cli.Run()
}
