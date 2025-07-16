package service

import (
	"math/rand/v2"
	"runtime"
	"sync"
	"time"
)

type Metrics struct {
	GaugeMetrics   (map[string]float64)
	CounterMetrics (map[string]int64)
	Mu             sync.Mutex
}

func NewMetrics() *Metrics {
	return &Metrics{
		GaugeMetrics:   make(map[string]float64),
		CounterMetrics: make(map[string]int64),
	}
}

func (m *Metrics) PollMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	m.Mu.Lock()
	defer m.Mu.Unlock()

	// Метрики gauge из runtime
	m.GaugeMetrics["Alloc"] = float64(memStats.Alloc)
	m.GaugeMetrics["BuckHashSys"] = float64(memStats.BuckHashSys)
	m.GaugeMetrics["Frees"] = float64(memStats.Frees)
	m.GaugeMetrics["GCCPUFraction"] = float64(memStats.GCCPUFraction)
	m.GaugeMetrics["GCSys"] = float64(memStats.GCSys)
	m.GaugeMetrics["HeapAlloc"] = float64(memStats.HeapAlloc)
	m.GaugeMetrics["HeapIdle"] = float64(memStats.HeapIdle)
	m.GaugeMetrics["HeapInuse"] = float64(memStats.HeapInuse)
	m.GaugeMetrics["HeapObjects"] = float64(memStats.HeapObjects)
	m.GaugeMetrics["HeapReleased"] = float64(memStats.HeapReleased)
	m.GaugeMetrics["HeapSys"] = float64(memStats.HeapSys)
	m.GaugeMetrics["LastGC"] = float64(memStats.LastGC)
	m.GaugeMetrics["Lookups"] = float64(memStats.Lookups)
	m.GaugeMetrics["MCacheInuse"] = float64(memStats.MCacheInuse)
	m.GaugeMetrics["MCacheSys"] = float64(memStats.MCacheSys)
	m.GaugeMetrics["MSpanInuse"] = float64(memStats.MSpanInuse)
	m.GaugeMetrics["MSpanSys"] = float64(memStats.MSpanSys)
	m.GaugeMetrics["Mallocs"] = float64(memStats.Mallocs)
	m.GaugeMetrics["NextGC"] = float64(memStats.NextGC)
	m.GaugeMetrics["NumForcedGC"] = float64(memStats.NumForcedGC)
	m.GaugeMetrics["NumGC"] = float64(memStats.NumGC)
	m.GaugeMetrics["OtherSys"] = float64(memStats.OtherSys)
	m.GaugeMetrics["PauseTotalNs"] = float64(memStats.PauseTotalNs)
	m.GaugeMetrics["StackInuse"] = float64(memStats.StackInuse)
	m.GaugeMetrics["StackSys"] = float64(memStats.StackSys)
	m.GaugeMetrics["Sys"] = float64(memStats.Sys)
	m.GaugeMetrics["TotalAlloc"] = float64(memStats.TotalAlloc)

	// Дополнительные метрики
	m.CounterMetrics["PollCount"]++
	m.GaugeMetrics["RandomValue"] = rand.Float64() * 100
}

func (m *Metrics) StartPolling(interval time.Duration, stopCh <-chan struct{}) {
	go func() {
		for {
			select {
			case <-stopCh:
				return
			default:
				m.PollMetrics()
				time.Sleep(interval)
			}
		}
	}()
}
