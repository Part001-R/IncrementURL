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

	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	baseAddrShortURL, serverAddr, err := config.ParseFlags()
	if err != nil {
		panic(err)
	}

	metrics := service.NewMetrics()
	metricsHandler := &handler.MetricsHandlerT{
		Metrics:          metrics,
		BaseAddrShortURL: config.Flags.FlagBaseAddrShortURL,
	}

	shortLong := service.NewShortByLong(baseAddrShortURL)
	shortLongHandler := &handler.ShortLongT{
		List:             shortLong,
		BaseAddrShortURL: baseAddrShortURL,
		ServerAddr:       serverAddr,
	}

	stopPolling := make(chan struct{})
	metrics.StartPolling(2*time.Second, stopPolling)

	cr := chi.NewRouter()
	cr.Post("/", shortLongHandler.ShortURLFromLong)
	cr.Get("/{id}", shortLongHandler.LongURLFromShort)
	cr.Get("/value/{type}/{name}", metricsHandler.DataMetricByTypeAndName)
	cr.Get("/", metricsHandler.AllMetricsHTML)

	fmt.Printf("Запуск сервера %s\n", config.Flags.FlagServerAddr)
	return http.ListenAndServe(config.Flags.FlagServerAddr, cr)
}
