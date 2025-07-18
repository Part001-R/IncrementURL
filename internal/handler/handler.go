package handler

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/Part001-R/IncrementURL/internal/service"
)

type MetricsHandlerT struct {
	Metrics *service.Metrics
}

type ShortLongT struct {
	List             *service.ShortByLong
	BaseAddrShortURL string
	ServerAddr       string
	mu               sync.Mutex
}

func (sl *ShortLongT) ShortURLFromLong(w http.ResponseWriter, r *http.Request) {

	sl.mu.Lock()
	defer sl.mu.Unlock()

	w.Header().Set("Content-Type", "text/plain")

	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	rxData, err := io.ReadAll(r.Body)
	defer func() {
		_ = r.Body.Close()
	}()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if len(rxData) == 0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	short := generateCode(8)
	sl.List.ListShorByLong[string(rxData)] = short

	strResult := "http://localhost" + sl.BaseAddrShortURL + short

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(strResult))

}

func (sl *ShortLongT) LongURLFromShort(w http.ResponseWriter, r *http.Request) {

	sl.mu.Lock()
	defer sl.mu.Unlock()

	w.Header().Set("Content-Type", "text/plain")

	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	rxData := r.URL.Path[1:]
	if len(rxData) == 0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	short := string(rxData)
	long := ""

	for k, v := range sl.List.ListShorByLong {
		if v == short {
			long = k
		}
	}
	if long == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", long)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (m *MetricsHandlerT) UpdateMetricByTypeAndName(w http.ResponseWriter, r *http.Request) {

	m.Metrics.Mu.Lock()
	defer m.Metrics.Mu.Unlock()

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
	m.Metrics.Mu.Lock()
	defer m.Metrics.Mu.Unlock()

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

// Генерация случайных символов заданной длинны
func generateCode(n int) string {
	b := make([]byte, n)
	io.ReadFull(rand.Reader, b)
	return base64.URLEncoding.EncodeToString(b)[:n]
}
