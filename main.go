package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/seniorcat/scraper/cmd"
	"github.com/seniorcat/scraper/pkg/metrics"
)

func startMetricsServer() {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}

func main() {
	metrics.Init()
	// Запуск HTTP-сервера для экспорта метрик
	go startMetricsServer()
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
