package handler

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type MetricsT struct {
	GaugeMetrics   map[string]float64
	CounterMetrics map[string]int64
	Mu             sync.RWMutex
}

func NewMetrics() *MetricsT {
	return &MetricsT{
		GaugeMetrics:   make(map[string]float64),
		CounterMetrics: make(map[string]int64),
		Mu:             sync.RWMutex{},
	}
}

type MetricsHandlerT struct {
	Metrics *MetricsT
}

type MetricsI interface {
	UpdateMetricByTypeAndName(w http.ResponseWriter, r *http.Request)
	AllMetricsHTML(w http.ResponseWriter, r *http.Request)
	ValueMetricByTypeAndName(w http.ResponseWriter, r *http.Request)
}

func NewMetricsStorage(m *MetricsT) MetricsI {
	return &MetricsHandlerT{
		Metrics: m,
	}
}

func (m *MetricsHandlerT) UpdateMetricByTypeAndName(w http.ResponseWriter, r *http.Request) {

	m.Metrics.Mu.RLock()
	defer m.Metrics.Mu.RUnlock()

	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	rxData := r.URL.Path[1:]
	slRxData := strings.Split(rxData, "/")
	if len(slRxData) != 4 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	typeMetric := slRxData[1] // gauge, counter
	nameMetric := slRxData[2]
	valueMetric := slRxData[3]

	if len(nameMetric) == 0 {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	switch typeMetric {
	case "counter":
		v, err := strconv.ParseInt(valueMetric, 10, 64)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		m.Metrics.CounterMetrics[nameMetric] += v

	case "gauge":
		v, err := strconv.ParseFloat(valueMetric, 64)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		m.Metrics.GaugeMetrics[nameMetric] = v

	default:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (m *MetricsHandlerT) AllMetricsHTML(w http.ResponseWriter, r *http.Request) {
	m.Metrics.Mu.RLock()
	defer m.Metrics.Mu.RUnlock()

	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintln(w, "<html><head><title>МЕТРИКИ</title></head><body>")
	fmt.Fprintln(w, "<h1>Доступные метрики</h1>")

	// Gauge метрики
	fmt.Fprintln(w, "<h2>Gauge</h2><ul>")
	gaugeKeys := make([]string, 0, len(m.Metrics.GaugeMetrics))
	for key := range m.Metrics.GaugeMetrics {
		gaugeKeys = append(gaugeKeys, key)
	}
	sort.Strings(gaugeKeys)
	for _, key := range gaugeKeys {
		fmt.Fprintf(w, "<li>%s: %f</li>\n", key, m.Metrics.GaugeMetrics[key])
	}
	fmt.Fprintln(w, "</ul>")

	// Counter метрики
	fmt.Fprintln(w, "<h2>Counter</h2><ul>")
	counterKeys := make([]string, 0, len(m.Metrics.CounterMetrics))
	for key := range m.Metrics.CounterMetrics {
		counterKeys = append(counterKeys, key)
	}
	sort.Strings(counterKeys)
	for _, key := range counterKeys {
		fmt.Fprintf(w, "<li>%s: %d</li>\n", key, m.Metrics.CounterMetrics[key])
	}
	fmt.Fprintln(w, "</ul>")
	fmt.Fprintln(w, "</body></html>")
}

func (m *MetricsHandlerT) ValueMetricByTypeAndName(w http.ResponseWriter, r *http.Request) {

	m.Metrics.Mu.RLock()
	defer m.Metrics.Mu.RUnlock()

	w.Header().Set("Content-Type", "text/plain")

	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	rxData := r.URL.Path[1:]
	slRxData := strings.Split(rxData, "/")
	metricType := slRxData[1]
	metricName := slRxData[2]

	val := ""

	switch metricType {
	case "counter":
		v, ok := m.Metrics.CounterMetrics[metricName]
		if !ok {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		val = fmt.Sprintf("%d", v)

	case "gauge":
		v, ok := m.Metrics.GaugeMetrics[metricName]
		if !ok {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		val = fmt.Sprintf("%f", v)

	default:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return

	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(val))
}
