package handler

import (
	"bytes"
	"fmt"
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

	shortLong := service.NewShortByLong(":8080")
	shortLongHandler := &ShortLongT{
		List: shortLong,
	}
	uLong := "https://practicum.yandex.ru/"
	bodyReq := bytes.NewBuffer([]byte(uLong))

	req := httptest.NewRequest(http.MethodPost, uLong, bodyReq)
	res := httptest.NewRecorder()

	shortLongHandler.ShortURLFromLong(res, req)

	resp := res.Result()
	defer func() {
		err := resp.Body.Close()
		assert.NoErrorf(t, err, "ошибка при закрытии потока {%v}", err)
	}()
	bodyResp, err := io.ReadAll(resp.Body)
	require.NoErrorf(t, err, "ошибка при чтении тела ответа:{%v}", err)

	assert.Equalf(t, http.StatusCreated, resp.StatusCode, "ожидался код {%d}, а принят {%d}", http.StatusCreated, resp.StatusCode)

	strRx := strings.Trim(string(bodyResp), "\"")
	_ = strRx
	strWant := fmt.Sprintf("http://localhost/%s", shortLongHandler.List.ListShorByLong[uLong])

	assert.Equalf(t, strWant, strRx, "ожидалось {%s} а принято {%s}", strWant, strRx)
}

func Test_ShortURLFromLong_FAULT(t *testing.T) {

	shortLong := &ShortLongT{
		List:             &service.ShortByLong{},
		BaseAddrShortURL: "http://localhost:8080/",
		ServerAddr:       "http://localhost:8080",
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
	}

	for _, tt := range testData {
		t.Run(tt.nameT, func(t *testing.T) {

			bodyReq := bytes.NewBuffer([]byte(tt.bodyT))

			req := httptest.NewRequest(tt.methodReqT, tt.urlT, bodyReq)
			res := httptest.NewRecorder()

			shortLong.ShortURLFromLong(res, req)

			resp := res.Result()
			defer func() {
				err := resp.Body.Close()
				assert.NoErrorf(t, err, "ошибка при закрытии потока {%v}", err)
			}()

			assert.Equalf(t, tt.wantStatusCode, resp.StatusCode, "ожидалcя код {%d}, а принят {%d}", tt.wantStatusCode, resp.StatusCode)
		})
	}
}

func Test_LongURLFromShort_SUCCESS(t *testing.T) {

	shortLong := service.NewShortByLong(":8080")
	shortLongHandler := &ShortLongT{
		List: shortLong,
	}

	uLong := "https://practicum.yandex.ru/"
	code := generateCode(6)
	shortLongHandler.List.ListShorByLong[uLong] = code

	urlReq := "http://localhost" + shortLongHandler.BaseAddrShortURL + "/" + code

	req := httptest.NewRequest(http.MethodGet, urlReq, nil)
	res := httptest.NewRecorder()

	shortLongHandler.LongURLFromShort(res, req)
	resp := res.Result()
	defer func() {
		err := resp.Body.Close()
		assert.NoErrorf(t, err, "ошибка при закрытии потока {%v}", err)
	}()

	rxHead := resp.Header.Get("Location")

	require.Equalf(t, http.StatusTemporaryRedirect, resp.StatusCode, "ожидался код {%d}, а принят {%d}", http.StatusTemporaryRedirect, resp.StatusCode)
	assert.Equalf(t, uLong, rxHead, "ожидался {%d}, а принято {%d}", uLong, rxHead)

}

func Test_LongURLFromShort_FAULT(t *testing.T) {

	shortLong := &ShortLongT{
		List:             &service.ShortByLong{},
		BaseAddrShortURL: "http://localhost:8080/",
		ServerAddr:       "http://localhost:8080",
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

			shortLong.LongURLFromShort(res, req)

			resp := res.Result()
			defer func() {
				err := resp.Body.Close()
				assert.NoErrorf(t, err, "ошибка при закрытии потока {%v}", err)
			}()

			assert.Equalf(t, tt.wantStatusCode, resp.StatusCode, "ожидался код {%d} а принят {%d}", tt.wantStatusCode, resp.StatusCode)
		})
	}
}

func Test_DataMetricByTypeAndName_SUCCESS(t *testing.T) {
	testMetrics := service.NewMetrics()

	testMetrics.CounterMetrics["PollCount"] = 123

	metricsHandler := &MetricsHandlerT{
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
			resp := res.Result()
			defer func() {
				err := resp.Body.Close()
				assert.NoErrorf(t, err, "ошибка при закрытии потока {%v}", err)
			}()

			require.Equalf(t, tt.wantStatusCode, resp.StatusCode, "ожидался код {%d}, а принят {%d}", tt.wantStatusCode, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoErrorf(t, err, "ошибка при чтении тела ответа {%v}", err)
			assert.Equalf(t, tt.wantBody, string(body), "принято{%s} а ожидалось {%s}", tt.wantBody, string(body))
		})
	}
}

func Test_DataMetricByTypeAndName_FAULT(t *testing.T) {
	testMetrics := service.NewMetrics()

	testMetrics.CounterMetrics["PollCount"] = 123

	metricsHandler := &MetricsHandlerT{
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
			resp := res.Result()
			defer func() {
				err := resp.Body.Close()
				assert.NoErrorf(t, err, "ошибка при закрытии потока {%v}", err)
			}()

			assert.Equalf(t, tt.wantStatusCode, resp.StatusCode, "ожидался код {%d}, а принят {%d}", tt.wantStatusCode, resp.StatusCode)
		})
	}
}

func TestAllMetricsHTML_SUCCESS(t *testing.T) {

	testMetrics := service.NewMetrics()

	testMetrics.CounterMetrics["PollCount"] = 10
	testMetrics.CounterMetrics["SomeCounter"] = 5
	testMetrics.GaugeMetrics["Alloc"] = 123.45
	testMetrics.GaugeMetrics["RandomValue"] = 987.65

	metricsHandler := &MetricsHandlerT{
		Metrics: testMetrics,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()

	metricsHandler.AllMetricsHTML(res, req)
	resp := res.Result()
	defer func() {
		err := resp.Body.Close()
		assert.NoErrorf(t, err, "ошибка при закрытии потока {%v}", err)
	}()

	require.Equalf(t, http.StatusOK, resp.StatusCode, "ожидался код ответа {%d}, а принят {%d}", http.StatusOK, resp.StatusCode)

	expectedContentType := "text/html; charset=utf-8"
	contentTypeRx := res.Header().Get("Content-Type")
	require.Equalf(t, expectedContentType, contentTypeRx, "ожидается контент {%s}, а принят {%s}", expectedContentType, contentTypeRx)

	body, err := io.ReadAll(resp.Body)
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

	metricsHandler := &MetricsHandlerT{
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
			resp := res.Result()
			defer func() {
				err := resp.Body.Close()
				assert.NoErrorf(t, err, "ошибка при закрытии потока {%v}", err)
			}()

			assert.Equalf(t, tt.wantStatusCode, resp.StatusCode, "ожидался код ответа {%d}, а принят {%d}", tt.wantStatusCode, resp.StatusCode)
		})
	}
}
