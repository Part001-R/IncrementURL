package service

import (
	"sync"
)

type Metrics struct {
	GaugeMetrics   map[string]float64
	CounterMetrics map[string]int64
	Mu             sync.Mutex
}

type ShortByLong struct {
	BaseAddrShortURL string
	ListShorByLong   map[string]string
}

func NewMetrics() *Metrics {
	return &Metrics{
		GaugeMetrics:   make(map[string]float64),
		CounterMetrics: make(map[string]int64),
	}
}

func NewShortByLong(baseURL string) *ShortByLong {
	return &ShortByLong{
		ListShorByLong:   make(map[string]string),
		BaseAddrShortURL: baseURL,
	}
}
