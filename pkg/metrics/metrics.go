package metrics

import "github.com/prometheus/client_golang/prometheus"

// Счетчик запросов парсера
var RequestCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "parser_requests_total",
		Help: "Total number of requests processed by the parser.",
	},
)

// Init регистрирует метрики
func Init() {
	prometheus.MustRegister(RequestCounter)
}
