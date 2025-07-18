package main

import (
	"fmt"
	"net/http"

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
		Metrics: metrics,
	}

	shortLong := service.NewShortByLong(baseAddrShortURL)
	shortLongHandler := &handler.ShortLongT{
		List:             shortLong,
		BaseAddrShortURL: baseAddrShortURL,
		ServerAddr:       serverAddr,
	}

	cr := chi.NewRouter()
	cr.Post("/", shortLongHandler.ShortURLFromLong)
	cr.Get("/{id}", shortLongHandler.LongURLFromShort)
	cr.Post("/update/{type}/{name}/{value}", metricsHandler.UpdateMetricByTypeAndName)
	cr.Get("/", metricsHandler.AllMetricsHTML)
	cr.Get("/value/{type}/{name}", metricsHandler.ValueMetricByTypeAndName)

	fmt.Printf("Запуск сервера %s\n", config.Flags.FlagServerAddr)
	return http.ListenAndServe(config.Flags.FlagServerAddr, cr)
}
