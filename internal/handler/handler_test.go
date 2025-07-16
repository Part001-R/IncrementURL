package handler

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Part001-R/IncrementURL/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ShortURLFromLong_SUCCESS(t *testing.T) {

	testMetrics := service.NewMetrics()

	metricsHandler := &MetricsHandler{
		Metrics:          testMetrics,
		BaseAddrShortURL: "http://localhost:8080/",
	}

	testData := []struct {
		nameT          string
		urlT           string
		methodReqT     string
		bodyT          string
		wantStatusCode int
		wantResultT    string
	}{
		{
			nameT:          "correct data",
			urlT:           "http://localhost:8080/",
			methodReqT:     http.MethodPost,
			bodyT:          "https://practicum.yandex.ru/",
			wantStatusCode: http.StatusCreated,
			wantResultT:    "http://localhost:8080/EwHXdJfB",
		},
	}

	for _, tt := range testData {
		t.Run(tt.nameT, func(t *testing.T) {

			bodyReq := bytes.NewBuffer([]byte(tt.bodyT))

			req := httptest.NewRequest(tt.methodReqT, tt.urlT, bodyReq)
			res := httptest.NewRecorder()

			metricsHandler.ShortURLFromLong(res, req)

			bodyResp, err := io.ReadAll(res.Body)
			require.NoErrorf(t, err, "ошибка при чтении тела ответа:{%v}", err)

			statusCodeRx := res.Result().StatusCode
			assert.Equalf(t, tt.wantStatusCode, statusCodeRx, "ожидался код {%d}, а принят {%d}", tt.wantStatusCode, statusCodeRx)

			strRx := strings.Trim(string(bodyResp), "\"")
			assert.Equalf(t, tt.wantResultT, strRx, "тело ответа {%s} не соответствует ожиданию {%s}", strRx, tt.wantResultT)
		})
	}
}

func Test_ShortURLFromLong_FAULT(t *testing.T) {
	testMetrics := service.NewMetrics()

	metricsHandler := &MetricsHandler{
		Metrics: testMetrics,
	}

	testData := []struct {
		nameT          string
		urlT           string
		methodReqT     string
		bodyT          string
		wantStatusCode int
	}{
		{
			nameT:          "wrong method",
			urlT:           "http://localhost:8080/",
			methodReqT:     http.MethodGet,
			bodyT:          "https://practicum.yandex.ru/",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			nameT:          "wrong url",
			urlT:           "http://localhost:8080/",
			methodReqT:     http.MethodPost,
			bodyT:          "https://practicum.google.ru/",
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range testData {
		t.Run(tt.nameT, func(t *testing.T) {

			bodyReq := bytes.NewBuffer([]byte(tt.bodyT))

			req := httptest.NewRequest(tt.methodReqT, tt.urlT, bodyReq)
			res := httptest.NewRecorder()

			metricsHandler.ShortURLFromLong(res, req)
			statusCodeRx := res.Result().StatusCode
			assert.Equalf(t, tt.wantStatusCode, statusCodeRx, "ожидалcя код {%d}, а принят {%d}", tt.wantStatusCode, statusCodeRx)
		})
	}

}

func Test_LongURLFromShort_SUCCESS(t *testing.T) {

	testMetrics := service.NewMetrics()

	metricsHandler := &MetricsHandler{
		Metrics: testMetrics,
	}

	testData := []struct {
		nameT          string
		urlT           string
		methodReqT     string
		wantStatusCode int
		wantResultT    string
	}{
		{
			nameT:          "correct data",
			urlT:           "http://localhost:8080/EwHXdJfB",
			methodReqT:     http.MethodGet,
			wantStatusCode: http.StatusTemporaryRedirect,
			wantResultT:    "Location: https://practicum.yandex.ru/",
		},
	}

	for _, tt := range testData {
		t.Run(tt.nameT, func(t *testing.T) {

			req := httptest.NewRequest(tt.methodReqT, tt.urlT, nil)
			res := httptest.NewRecorder()

			metricsHandler.LongURLFromShort(res, req)

			dataRx, err := io.ReadAll(res.Body)
			require.NoErrorf(t, err, "ошибка при чтении тела ответа {%v}", err)

			statusCodeRx := res.Result().StatusCode
			require.Equalf(t, tt.wantStatusCode, statusCodeRx, "ожидался {%d}, а принято {%d}", tt.wantStatusCode, statusCodeRx)

			strRx := string(dataRx)
			assert.Equalf(t, tt.wantResultT, strRx, "ожидалось {%s}, а принято {%s}", tt.wantResultT, strRx)
		})
	}
}

func Test_LongURLFromShort_FAULT(t *testing.T) {
	testMetrics := service.NewMetrics()

	metricsHandler := &MetricsHandler{
		Metrics: testMetrics,
	}

	testData := []struct {
		nameT          string
		urlT           string
		methodReqT     string
		wantStatusCode int
	}{
		{
			nameT:          "wrong method",
			urlT:           "http://localhost:8080/EwHXdJfB",
			methodReqT:     http.MethodPost,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			nameT:          "missing {id}",
			urlT:           "http://localhost:8080/",
			methodReqT:     http.MethodGet,
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range testData {
		t.Run(tt.nameT, func(t *testing.T) {

			req := httptest.NewRequest(tt.methodReqT, tt.urlT, nil)
			res := httptest.NewRecorder()

			metricsHandler.LongURLFromShort(res, req)
			statusCodeRx := res.Result().StatusCode

			assert.Equalf(t, tt.wantStatusCode, statusCodeRx, "ожидался код {%d} а принят {%d}", tt.wantStatusCode, statusCodeRx)
		})
	}
}

func Test_DataMetricByTypeAndName_SUCCESS(t *testing.T) {
	testMetrics := service.NewMetrics()

	testMetrics.CounterMetrics["PollCount"] = 123

	metricsHandler := &MetricsHandler{
		Metrics: testMetrics,
	}
	testsData := []struct {
		nameT          string
		methodT        string
		urlT           string
		wantStatusCode int
		wantBody       string
	}{
		{
			nameT:          "correct data",
			methodT:        http.MethodGet,
			urlT:           "http://localhost:8080/value/counter/PollCount",
			wantStatusCode: http.StatusOK,
			wantBody:       "123",
		},
	}
	for _, tt := range testsData {
		t.Run(tt.nameT, func(t *testing.T) {
			req := httptest.NewRequest(tt.methodT, tt.urlT, nil)
			res := httptest.NewRecorder()

			metricsHandler.DataMetricByTypeAndName(res, req)
			require.Equalf(t, tt.wantStatusCode, res.Result().StatusCode, "ожидался код {%d}, а принят {%d}", tt.wantStatusCode, res.Result().StatusCode)

			body, err := io.ReadAll(res.Body)
			require.NoErrorf(t, err, "ошибка при чтении тела ответа {%v}", err)
			assert.Equalf(t, tt.wantBody, string(body), "принято{%s} а ожидалось {%s}", tt.wantBody, string(body))
		})
	}
}

func Test_DataMetricByTypeAndName_FAULT(t *testing.T) {
	testMetrics := service.NewMetrics()

	testMetrics.CounterMetrics["PollCount"] = 123

	metricsHandler := &MetricsHandler{
		Metrics: testMetrics,
	}
	testsData := []struct {
		nameT          string
		methodT        string
		urlT           string
		wantStatusCode int
		wantBody       string
	}{
		{
			nameT:          "wrong metric type",
			methodT:        http.MethodGet,
			urlT:           "http://localhost:8080/value/wrong/PollCount",
			wantStatusCode: http.StatusNotFound,
		},
		{
			nameT:          "wrong metric name",
			methodT:        http.MethodGet,
			urlT:           "http://localhost:8080/value/counter/wrong",
			wantStatusCode: http.StatusNotFound,
		},
		{
			nameT:          "wrong URL",
			methodT:        http.MethodGet,
			urlT:           "http://localhost:8080/value/counter",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			nameT:          "wrong method",
			methodT:        http.MethodPost,
			urlT:           "http://localhost:8080/value/counter",
			wantStatusCode: http.StatusBadRequest,
		},
	}
	for _, tt := range testsData {
		t.Run(tt.nameT, func(t *testing.T) {
			req := httptest.NewRequest(tt.methodT, tt.urlT, nil)
			res := httptest.NewRecorder()

			metricsHandler.DataMetricByTypeAndName(res, req)
			assert.Equalf(t, tt.wantStatusCode, res.Result().StatusCode, "ожидался код {%d}, а принят {%d}", tt.wantStatusCode, res.Result().StatusCode)
		})
	}
}

func TestAllMetricsHTML_SUCCESS(t *testing.T) {

	testMetrics := service.NewMetrics()

	testMetrics.CounterMetrics["PollCount"] = 10
	testMetrics.CounterMetrics["SomeCounter"] = 5
	testMetrics.GaugeMetrics["Alloc"] = 123.45
	testMetrics.GaugeMetrics["RandomValue"] = 987.65

	metricsHandler := &MetricsHandler{
		Metrics: testMetrics,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()

	metricsHandler.AllMetricsHTML(res, req)
	require.Equalf(t, http.StatusOK, res.Code, "ожидался код ответа {%d}, а принят {%d}", http.StatusOK, res.Code)

	expectedContentType := "text/html; charset=utf-8"
	contentTypeRx := res.Header().Get("Content-Type")
	require.Equalf(t, expectedContentType, contentTypeRx, "ожидается контент {%s}, а принят {%s}", expectedContentType, contentTypeRx)

	body, err := io.ReadAll(res.Body)
	require.NoErrorf(t, err, "ошибка при чтении тела ответа {%v}", err)

	responseBody := string(body)

	assert.Contains(t, responseBody, "<html>", "HTML нет <html> тега")
	assert.Contains(t, responseBody, "<title>МЕТРИКИ</title>", "HTML нет title")
	assert.Contains(t, responseBody, "<h1>Доступные метрики</h1>", "HTML нет h1 заголовка")
	assert.Contains(t, responseBody, "<h2>Gauge</h2>", "HTML нет Gauge заголовка")
	assert.Contains(t, responseBody, "<h2>Counter</h2>", "HTML нет Counter заголовка")

	assert.Contains(t, responseBody, "<li>Alloc: 123.450000</li>", "HTML нет соответствия Alloc")
	assert.Contains(t, responseBody, "<li>PollCount: 10</li>", "HTML нет соответствия PollCount")
	assert.Contains(t, responseBody, "<li>SomeCounter: 5</li>", "HTML нет соответствия SomeCounter")

	assert.Contains(t, responseBody, "<li>RandomValue: 987.650000</li>", "HTML нет соответствия RandomValue")
	assert.Contains(t, responseBody, "</li>", "HTML нет тега заурытия (</li>)")

}

func TestAllMetricsHTML_FAULT(t *testing.T) {

	testMetrics := service.NewMetrics()

	metricsHandler := &MetricsHandler{
		Metrics: testMetrics,
	}

	testsData := []struct {
		nameT          string
		methodT        string
		urlT           string
		wantStatusCode int
	}{
		{
			nameT:          "wrong method",
			methodT:        http.MethodPost,
			urlT:           "http://localhost:8080/",
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range testsData {
		t.Run(tt.nameT, func(t *testing.T) {
			req := httptest.NewRequest(tt.methodT, tt.urlT, nil)
			res := httptest.NewRecorder()

			metricsHandler.AllMetricsHTML(res, req)
			assert.Equalf(t, tt.wantStatusCode, res.Code, "ожидался код ответа {%d}, а принят {%d}", tt.wantStatusCode, res.Code)

		})
	}
}
