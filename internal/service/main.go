package service

import (
	"sync"
)

type Metrics struct {
	GaugeMetrics   map[string]float64
	CounterMetrics map[string]int64
	Mu             sync.RWMutex
}

type ShortLongURL struct {
	BaseAddrShortURL string
	ShorByLong       map[string]string
	LongByShort      map[string]string
}

func NewMetrics() *Metrics {
	return &Metrics{
		GaugeMetrics:   make(map[string]float64),
		CounterMetrics: make(map[string]int64),
	}
}

func NewShortLongURL(baseURL string) *ShortLongURL {
	return &ShortLongURL{
		ShorByLong:       make(map[string]string),
		LongByShort:      make(map[string]string),
		BaseAddrShortURL: baseURL,
	}
}
