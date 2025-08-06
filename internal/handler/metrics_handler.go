package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/Part001-R/IncrementURL/internal/config/config"
)

type MetricsT struct {
	GaugeMetrics   map[string]float64
	CounterMetrics map[string]int64
	Mu             sync.RWMutex
}

type MetricsHandlerT struct {
	Metrics             *MetricsT
	StoreIntervalMetr   string
	FileStoragePathMetr string
	RestoreMetr         string
}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

type EventMetricT struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

type MetricsI interface {
	UpdateMetricByTypeAndName(w http.ResponseWriter, r *http.Request)
	AllMetricsHTML(w http.ResponseWriter, r *http.Request)
	ValueMetricByTypeAndName(w http.ResponseWriter, r *http.Request)
	MetricByJSON(w http.ResponseWriter, r *http.Request)
	StorageMetrics() error
	LoadFileMetrics() error
}

func NewMetricsStorage(m *MetricsT, f config.FlagsT) MetricsI {
	return &MetricsHandlerT{
		Metrics:             m,
		StoreIntervalMetr:   f.StoreIntervalMetr,
		FileStoragePathMetr: f.FileStoragePathMetr,
		RestoreMetr:         f.RestoreMetr,
	}
}

func NewMetrics() *MetricsT {
	return &MetricsT{
		GaugeMetrics:   make(map[string]float64),
		CounterMetrics: make(map[string]int64),
		Mu:             sync.RWMutex{},
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

	if m.StoreIntervalMetr == "0" { // Синхронное сохранение
		err := storage(m.FileStoragePathMetr, m.Metrics.GaugeMetrics, m.Metrics.CounterMetrics)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (m *MetricsHandlerT) MetricByJSON(w http.ResponseWriter, r *http.Request) {
	m.Metrics.Mu.RLock()
	defer m.Metrics.Mu.RUnlock()

	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if r.Header.Get("Content-Type") != `application/json` {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var rxData Metrics
	err := json.NewDecoder(r.Body).Decode(&rxData)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if rxData.ID == "" || rxData.MType == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	rawData := Metrics{}

	switch rxData.MType {
	case "counter":
		if _, exists := m.Metrics.CounterMetrics[rxData.ID]; !exists {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		value := m.Metrics.CounterMetrics[rxData.ID]
		rawData.Delta = &value

	case "gauge":
		if _, exists := m.Metrics.GaugeMetrics[rxData.ID]; !exists {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		value := m.Metrics.GaugeMetrics[rxData.ID]
		rawData.Value = &value

	default:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	rawData.ID = rxData.ID
	rawData.MType = rxData.MType

	txData, err := json.Marshal(rawData)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(txData)
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

func (m *MetricsHandlerT) StorageMetrics() error {

	m.Metrics.Mu.RLock()
	defer m.Metrics.Mu.RUnlock()

	err := storage(m.FileStoragePathMetr, m.Metrics.GaugeMetrics, m.Metrics.CounterMetrics)
	if err != nil {
		return fmt.Errorf("ошибка сохранения значений метрик в файл <%w>", err)
	}

	return nil
}

func (m *MetricsHandlerT) LoadFileMetrics() error {

	m.Metrics.Mu.RLock()
	defer m.Metrics.Mu.RUnlock()

	// Проверка
	if m.FileStoragePathMetr == "" {
		return errors.New("не указан путь к файлу с метриками")
	}
	if m.Metrics.CounterMetrics == nil {
		return fmt.Errorf("нет указателя на мапу метрик counter")
	}
	if m.Metrics.GaugeMetrics == nil {
		return fmt.Errorf("нет указателя на мапу метрик counter")
	}

	// Файл
	file, err := os.OpenFile(m.FileStoragePathMetr, os.O_RDONLY, 0644)
	if err != nil {
		return fmt.Errorf("ошибка открытия файла <%s>: %v", m.FileStoragePathMetr, err)
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		return fmt.Errorf("ошибка получения статуса файла: %v", err)
	}
	if fi.Size() == 0 {
		return nil
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("ошибка чтения файла: %v", err)
	}

	var events []EventMetricT
	if err := json.Unmarshal(data, &events); err != nil {
		return fmt.Errorf("ошибка Unmarshal: %v", err)
	}

	// Сохранение данных из файла в мапы
	for _, ev := range events {

		if ev.Type == "counter" {
			value, err := strconv.ParseInt(ev.Value, 10, 64)
			if err != nil {
				return fmt.Errorf("ошибка парсинга <%s>, c значением <%s>", ev.ID, ev.Value)
			}
			m.Metrics.CounterMetrics[ev.ID] = value
		}

		if ev.Type == "gauge" {
			value, err := strconv.ParseFloat(ev.Value, 64)
			if err != nil {
				return fmt.Errorf("ошибка парсинга <%s>, c значением <%s>", ev.ID, ev.Value)
			}
			m.Metrics.GaugeMetrics[ev.ID] = value
		}
	}
	return nil
}

func storage(filePath string, gM map[string]float64, cM map[string]int64) error {

	// Проверка аргументов
	if filePath == "" {
		return errors.New("не указан путь к файлу хранения метрик")
	}
	if gM == nil {
		return errors.New("нет указателя на мапу gauge метрик")
	}
	if cM == nil {
		return errors.New("нет указателя на мапу counter метрик")
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("ошибка <%v> открытия файла <%s>", err, filePath)
	}
	defer func() {
		_ = file.Close()
	}()

	if len(gM) == 0 && len(cM) == 0 {
		return nil
	}

	// Вывод с построением строк
	file.WriteString("[\n")

	// Обработка gauge
	if len(gM) != 0 {
		numb := 1
		size := len(gM)
		for k, v := range gM {

			str := fmt.Sprintf(`	{"id":"%s","type":"%s","value":"%.f"}`, k, "gauge", v)
			file.WriteString(str)
			if numb < size {
				file.WriteString(",\n")
			}
			numb++
		}
	}
	// Обработка counter
	if len(cM) != 0 {

		if len(gM) > 0 {
			file.WriteString(",\n")
		}

		numb := 1
		size := len(cM)
		for k, v := range cM {

			str := fmt.Sprintf(`	{"id":"%s","type":"%s","delta":"%d"}`, k, "counter", v)
			file.WriteString(str)
			if numb < size {
				file.WriteString(",\n")
			}
			numb++
		}
	}

	file.WriteString("\n")
	file.WriteString("]\n")

	return nil
}
