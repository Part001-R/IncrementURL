package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ShortURLFromLong_SUCCESS(t *testing.T) {

	shortLong := NewShortLongURL()
	shortLongHandler := &ShortLongT{
		List: shortLong,
	}

	testData := []struct {
		nameT          string
		urlT           string
		methodReqT     string
		bodyT          string
		wantStatusCode int
	}{
		{
			nameT:          "correct data",
			urlT:           "http://localhost:8080",
			methodReqT:     http.MethodPost,
			bodyT:          "https://practicum.yandex.ru/",
			wantStatusCode: http.StatusCreated,
		},
	}

	for _, tt := range testData {
		t.Run(tt.nameT, func(t *testing.T) {

			bodyReq := bytes.NewBuffer([]byte(tt.bodyT))

			req := httptest.NewRequest(tt.methodReqT, tt.urlT, bodyReq)
			res := httptest.NewRecorder()
			shortLongHandler.ShortURLFromLong(res, req)

			resp := res.Result()
			defer func() {
				err := resp.Body.Close()
				assert.NoErrorf(t, err, "ошибка закрытия потока {%v}", err)
			}()

			require.Equalf(t, tt.wantStatusCode, resp.StatusCode, "ожидался код {%d}, а принят {%d}", tt.wantStatusCode, resp.StatusCode)

			_, ok := shortLongHandler.List.ShorByLong[tt.bodyT]
			assert.Equalf(t, ok, true, "нет признака существования ключа в мапе")
		})
	}
}

func Test_ShortURLFromLong_FAULT(t *testing.T) {

	shortLong := &ShortLongT{
		List:             &ShortLongUrlT{},
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

	shortLong := NewShortLongURL()
	shortLongHandler := &ShortLongT{
		List: shortLong,
	}
	shortLongHandler.List.LongByShort = make(map[string]string)

	uLong := "https://practicum.yandex.ru/"

	code, err := generateCode(6)
	require.NoErrorf(t, err, "ожидалось отсутствие ошибки, а принято {%v}", err)

	shortLongHandler.List.LongByShort[code] = uLong

	urlReq := fmt.Sprintf("http://localhost:8080/%s", code)
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
	assert.Equalf(t, uLong, rxHead, "ожидался {%s}, а принято {%s}", uLong, rxHead)
}

func Test_LongURLFromShort_FAULT(t *testing.T) {

	shortLong := &ShortLongT{
		List:             &ShortLongUrlT{},
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

func Test_UpdateMetricByTypeAndName_SUCCESS(t *testing.T) {
	testMetrics := NewMetrics()

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
			nameT:          "correct data",
			methodT:        http.MethodPost,
			urlT:           "http://localhost:8080/update/counter/PollCount/222",
			wantStatusCode: http.StatusOK,
		},
	}
	for _, tt := range testsData {
		t.Run(tt.nameT, func(t *testing.T) {
			req := httptest.NewRequest(tt.methodT, tt.urlT, nil)
			res := httptest.NewRecorder()

			metricsHandler.UpdateMetricByTypeAndName(res, req)
			resp := res.Result()
			defer func() {
				err := resp.Body.Close()
				assert.NoErrorf(t, err, "ошибка при закрытии потока {%v}", err)
			}()

			require.Equalf(t, tt.wantStatusCode, resp.StatusCode, "ожидался код {%d}, а принят {%d}", tt.wantStatusCode, resp.StatusCode)
		})
	}
}

func Test_UpdateMetricByTypeAndName_FAULT(t *testing.T) {
	testMetrics := NewMetrics()

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
			methodT:        http.MethodPost,
			urlT:           "http://localhost:8080/update/wrong/PollCount/1",
			wantStatusCode: http.StatusNotFound,
		},
		{
			nameT:          "wrong URL",
			methodT:        http.MethodPost,
			urlT:           "http://localhost:8080/update/counter//1",
			wantStatusCode: http.StatusNotFound,
		},
		{
			nameT:          "wrong method",
			methodT:        http.MethodGet,
			urlT:           "http://localhost:8080/update/gauge/PollCount/1",
			wantStatusCode: http.StatusBadRequest,
		},
	}
	for _, tt := range testsData {
		t.Run(tt.nameT, func(t *testing.T) {
			req := httptest.NewRequest(tt.methodT, tt.urlT, nil)
			res := httptest.NewRecorder()

			metricsHandler.UpdateMetricByTypeAndName(res, req)
			resp := res.Result()
			defer func() {
				err := resp.Body.Close()
				assert.NoErrorf(t, err, "ошибка при закрытии потока {%v}", err)
			}()

			assert.Equalf(t, tt.wantStatusCode, resp.StatusCode, "ожидался код {%d}, а принят {%d}", tt.wantStatusCode, resp.StatusCode)
		})
	}
}

func Test_ValueMetricByTypeAndName_SUCCESS(t *testing.T) {
	testMetrics := NewMetrics()

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

			metricsHandler.ValueMetricByTypeAndName(res, req)
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

func Test_ValueMetricByTypeAndName_FAULT(t *testing.T) {
	testMetrics := NewMetrics()

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
			nameT:          "wrong URL",
			methodT:        http.MethodGet,
			urlT:           "http://localhost:8080/value/counter//",
			wantStatusCode: http.StatusNotFound,
		},
		{
			nameT:          "wrong method",
			methodT:        http.MethodPost,
			urlT:           "http://localhost:8080/value/gauge/PollCount",
			wantStatusCode: http.StatusBadRequest,
		},
	}
	for _, tt := range testsData {
		t.Run(tt.nameT, func(t *testing.T) {
			req := httptest.NewRequest(tt.methodT, tt.urlT, nil)
			res := httptest.NewRecorder()

			metricsHandler.ValueMetricByTypeAndName(res, req)
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

	testMetrics := NewMetrics()

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

	testMetrics := NewMetrics()

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
