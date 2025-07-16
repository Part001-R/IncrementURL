package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Part001-R/IncrementURL/internal/config/config"
	"github.com/Part001-R/IncrementURL/internal/handler"
	"github.com/Part001-R/IncrementURL/internal/service"
	"github.com/go-chi/chi/v5"
)

func main() {

	err := config.ParseFlags()
	if err != nil {
		panic(err)
	}

	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	metrics := service.NewMetrics()
	metricsHandler := &handler.MetricsHandler{
		Metrics:          metrics,
		BaseAddrShortURL: config.Flags.FlagBaseAddrShortURL,
	}

	stopPolling := make(chan struct{})
	metrics.StartPolling(2*time.Second, stopPolling)

	cr := chi.NewRouter()
	cr.Post("/", metricsHandler.ShortURLFromLong)
	cr.Get("/{id}", metricsHandler.LongURLFromShort)
	cr.Get("/value/{type}/{name}", metricsHandler.DataMetricByTypeAndName)
	cr.Get("/", metricsHandler.AllMetricsHTML)

	fmt.Printf("Запуск сервера на порту:%s\n", config.Flags.FlagServerAddr)
	return http.ListenAndServe(config.Flags.FlagServerAddr, cr)
}
