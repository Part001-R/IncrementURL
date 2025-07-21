package service

import (
	"fmt"
	"net/http"

	"github.com/Part001-R/IncrementURL/internal/config/config"
	"github.com/Part001-R/IncrementURL/internal/handler"
	"github.com/go-chi/chi"
)

func Run() error {
	baseAddrShortURL, serverAddr, err := config.ParseFlags()
	if err != nil {
		return fmt.Errorf("ошибка чтения флагов: {%w}", err)
	}

	metrics := handler.NewMetrics()
	storageMetrics := handler.NewMetricsStorage(metrics)

	shortLong := handler.NewShortLongURL()
	storageLongShort := handler.NewShortLongStorage(shortLong, baseAddrShortURL, serverAddr)

	cr := chi.NewRouter()
	cr.Post("/", storageLongShort.ShortURLFromLong)
	cr.Get("/{id}", storageLongShort.LongURLFromShort)
	cr.Post("/update/{type}/{name}/{value}", storageMetrics.UpdateMetricByTypeAndName)
	cr.Get("/", storageMetrics.AllMetricsHTML)
	cr.Get("/value/{type}/{name}", storageMetrics.ValueMetricByTypeAndName)

	fmt.Printf("Запуск сервера %s\n", serverAddr)
	err = http.ListenAndServe(serverAddr, cr)
	return fmt.Errorf("ошибка http сервера: {%w}", err)
}
