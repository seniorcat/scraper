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

	// Регистрация команды "initdb" для создания таблиц
	cli.RegisterCommand("initdb", "Создание таблиц в базе данных", func(args []string) {
		cmd.CreateTables()
	})

	// Добавляем команду "help" для справки
	cli.RegisterCommand("help", "Вывод справки по командам", func(args []string) {
		cli.PrintHelp()
	})

	// Запуск CLI
	cli.Run()
}
