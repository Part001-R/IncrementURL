package service

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Part001-R/IncrementURL/internal/config/config"
	"github.com/Part001-R/IncrementURL/internal/handler"
	"github.com/Part001-R/IncrementURL/internal/service/logger"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

type paramsURLT struct {
	flags            config.FlagsT
	storageMetrics   handler.MetricsI
	storageLongShort handler.ShortLongI
}

func Run() error {

	params, err := prepare()
	if err != nil {
		return fmt.Errorf("ошибка подготовительных действий <%w>", err)
	}

	err = server(params)
	if err != nil {
		return fmt.Errorf("ошибка сервера <%w>", err)
	}

	return nil
}

func prepare() (*paramsURLT, error) {

	// Флаги
	flags := config.ParseFlags()

	// Логер
	err := logger.Initialize(flags.LogLevel)
	if err != nil {
		return &paramsURLT{}, fmt.Errorf("ошибка инициализации логера <%w>", err)
	}

	// Метрики
	metrics := handler.NewMetrics()
	storageMetrics := handler.NewMetricsStorage(metrics, flags)

	if flags.RestoreMetr == "true" {
		err := storageMetrics.LoadFileMetrics()
		if err != nil {
			return &paramsURLT{}, fmt.Errorf("ошибка чтения файла метрик при запуске <%w>", err)
		}
	}

	// Ссылки
	shortLong := handler.NewShortLongURL()
	storageLongShort := handler.NewShortLongStorage(shortLong, flags)
	err = storageLongShort.LoadFileURL()
	if err != nil {
		return &paramsURLT{}, fmt.Errorf("ошибка загрузки данных файла <%w>", err)
	}

	return &paramsURLT{
		flags:            flags,
		storageMetrics:   storageMetrics,
		storageLongShort: storageLongShort,
	}, nil
}

func server(params *paramsURLT) error {

	cr := chi.NewRouter()

	// Ссылки
	cr.Post("/", handler.Middleware(http.HandlerFunc(params.storageLongShort.ShortURLFromLong)))
	cr.Post("/api/shorten", handler.Middleware(http.HandlerFunc(params.storageLongShort.ShortURLFromLongJSON)))
	cr.Get("/{id}", handler.Middleware(http.HandlerFunc(params.storageLongShort.LongURLFromShort)))
	// Метрики
	cr.Post("/update/{type}/{name}/{value}", handler.Middleware(http.HandlerFunc(params.storageMetrics.UpdateMetricByTypeAndName)))
	cr.Post("/update", handler.Middleware(http.HandlerFunc(params.storageMetrics.MetricByJSON)))
	cr.Get("/", handler.Middleware(http.HandlerFunc(params.storageMetrics.AllMetricsHTML)))
	cr.Get("/value/{type}/{name}", handler.Middleware(http.HandlerFunc(params.storageMetrics.ValueMetricByTypeAndName)))

	/*
		logger.Log.Info("Запуск сервера", zap.String("address", params.flags.ServerAddr))
		err := http.ListenAndServe(params.flags.ServerAddr, cr)
		return fmt.Errorf("ошибка http сервера: <%w>", err)
	*/

	srvConf := &http.Server{
		Addr:    params.flags.ServerAddr,
		Handler: cr,
	}

	// Запуск сервера
	logger.Log.Info("Запуск сервера", zap.String("address", srvConf.Addr))
	chSrvErr := make(chan error)

	go func(srv *http.Server, txErr chan error) {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Log.Error("Ошибка при запуске сервера", zap.Error(err))
		}
		txErr <- err
	}(srvConf, chSrvErr)

	// Обработка периодического сохранения метрик
	chStorageErr := make(chan error)
	if params.flags.StoreIntervalMetr != "0" {

		periodSec, err := strconv.Atoi(params.flags.StoreIntervalMetr)
		if err != nil {
			return fmt.Errorf("ошибка в формате представления периода сохранения метрик <%w>", err)
		}

		ticker := time.NewTicker(time.Duration(periodSec) * time.Second)
		defer ticker.Stop()

		go func(txErr chan error) {
			for range ticker.C {
				if err := params.storageMetrics.StorageMetrics(); err != nil {
					txErr <- err
				}
			}
		}(chStorageErr)
	}

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// Сигналы остановки
	select {
	case <-sigchan:
		err := params.storageMetrics.StorageMetrics()
		if err != nil {
			logger.Log.Error("ошибка сохранения метрик при штатном завершении работы", zap.String("ошибка", err.Error()))
		}
		logger.Log.Info("сервер остановлен штатно", zap.String("address", srvConf.Addr))
		return nil
	case err := <-chSrvErr:
		logger.Log.Error("ошибка сервера", zap.String("address", srvConf.Addr), zap.String("ошибка", err.Error()))
		return err
	case err := <-chStorageErr:
		logger.Log.Error("ошибка периодического сохранения метрик в файл", zap.String("address", srvConf.Addr), zap.String("ошибка", err.Error()))
		return err
	}
}
