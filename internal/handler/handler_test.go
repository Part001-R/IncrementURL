package handler

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/Part001-R/IncrementURL/internal/config/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// URL

func Test_ShortURLFromLong_SUCCESS(t *testing.T) {

	//flags := config.ParseFlags()

	shortLong := NewShortLongURL()
	shortLongHandler := &ShortLongT{
		List:             shortLong,
		BaseAddrShortURL: ":8080/",
		ServerAddr:       ":8080",
		FileStoragePath:  "storage.json",
	}

	testData := []struct {
		nameT          string
		urlT           string
		methodReqT     string
		bodyT          string
		wantStatusCode int
	}{
		{
			nameT:          "запись 1",
			urlT:           "http://localhost:8080",
			methodReqT:     http.MethodPost,
			bodyT:          "https://practicum.yandex.ru/",
			wantStatusCode: http.StatusCreated,
		},
		{
			nameT:          "запись 2",
			urlT:           "http://localhost:8080",
			methodReqT:     http.MethodPost,
			bodyT:          "https://AAA/",
			wantStatusCode: http.StatusCreated,
		},
		{
			nameT:          "запись 3",
			urlT:           "http://localhost:8080",
			methodReqT:     http.MethodPost,
			bodyT:          "https://BBB/",
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

	err := os.Remove(shortLongHandler.FileStoragePath)
	assert.NoErrorf(t, err, "неожиданная ошибка при удалении файла <%v>", err)
}

func Test_ShortURLFromLong_FAULT(t *testing.T) {

	shortLong := &ShortLongT{
		List:             &ShortLongURLT{},
		BaseAddrShortURL: ":8080/",
		ServerAddr:       ":8080",
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

func Test_ShortURLFromLongJSON_SUCCESS(t *testing.T) {

	shortLong := NewShortLongURL()
	shortLongHandler := &ShortLongT{
		List: shortLong,
	}

	testData := []struct {
		nameT          string
		urlT           string
		methodReqT     string
		bodyT          rxLongURLT
		wantStatusCode int
	}{
		{
			nameT:          "correct data",
			urlT:           "http://localhost:8080/api/shorten",
			methodReqT:     http.MethodPost,
			bodyT:          rxLongURLT{URL: "https://practicum.yandex.ru"},
			wantStatusCode: http.StatusCreated,
		},
	}

	for _, tt := range testData {
		t.Run(tt.nameT, func(t *testing.T) {

			txData, err := json.Marshal(tt.bodyT)
			require.NoErrorf(t, err, "ожидалось отсутствие ошибка при маршалинге, а принято <%v>", err)

			bodyReq := bytes.NewBuffer([]byte(txData))

			req := httptest.NewRequest(tt.methodReqT, tt.urlT, bodyReq)
			res := httptest.NewRecorder()

			req.Header.Set("content-Type", "application/json")
			shortLongHandler.ShortURLFromLongJSON(res, req)

			resp := res.Result()
			defer func() {
				err := resp.Body.Close()
				assert.NoErrorf(t, err, "ошибка закрытия потока {%v}", err)
			}()

			require.Equalf(t, tt.wantStatusCode, resp.StatusCode, "ожидался код {%d}, а принят {%d}", tt.wantStatusCode, resp.StatusCode)

			_, ok := shortLongHandler.List.ShorByLong[tt.bodyT.URL]
			assert.Equalf(t, ok, true, "нет признака существования ключа в мапе")
		})
	}
}

func Test_ShortURLFromLongJSON_FAULT(t *testing.T) {

	shortLong := NewShortLongURL()
	shortLongHandler := &ShortLongT{
		List: shortLong,
	}

	testData := []struct {
		nameT          string
		urlT           string
		methodReqT     string
		bodyT          rxLongURLT
		contentType    string
		wantStatusCode int
	}{
		{
			nameT:          "неподдерживаемый метод",
			urlT:           "http://localhost:8080/api/shorten",
			methodReqT:     http.MethodGet,
			bodyT:          rxLongURLT{URL: "https://practicum.yandex.ru"},
			contentType:    "application/json",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			nameT:          "нет данных URL",
			urlT:           "http://localhost:8080/api/shorten",
			methodReqT:     http.MethodPost,
			bodyT:          rxLongURLT{URL: ""},
			contentType:    "application/json",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			nameT:          "неподдерживаемый тип контента",
			urlT:           "http://localhost:8080/api/shorten",
			methodReqT:     http.MethodPost,
			bodyT:          rxLongURLT{URL: "https://practicum.yandex.ru"},
			contentType:    "application/AAA",
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range testData {
		t.Run(tt.nameT, func(t *testing.T) {

			txData, err := json.Marshal(tt.bodyT)
			require.NoErrorf(t, err, "ожидалось отсутствие ошибка при маршалинге, а принято <%v>", err)

			bodyReq := bytes.NewBuffer([]byte(txData))

			req := httptest.NewRequest(tt.methodReqT, tt.urlT, bodyReq)
			res := httptest.NewRecorder()
			shortLongHandler.ShortURLFromLongJSON(res, req)

			resp := res.Result()
			defer func() {
				err := resp.Body.Close()
				assert.NoErrorf(t, err, "ошибка закрытия потока {%v}", err)
			}()

			require.Equalf(t, tt.wantStatusCode, resp.StatusCode, "ожидался код {%d}, а принят {%d}", tt.wantStatusCode, resp.StatusCode)
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
		List:             &ShortLongURLT{},
		BaseAddrShortURL: ":8080/",
		ServerAddr:       ":8080",
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

func Test_Middleware_SUCCESS(t *testing.T) {

	//flags := config.ParseFlags()

	shortLong := NewShortLongURL()
	shortLongHandler := &ShortLongT{
		List:             shortLong,
		BaseAddrShortURL: ":8080/",
		ServerAddr:       ":8080",
		FileStoragePath:  "storage.json",
	}

	handler := http.HandlerFunc(Middleware(shortLongHandler.ShortURLFromLong))

	srv := httptest.NewServer(handler)
	defer srv.Close()

	testData := []struct {
		nameT            string
		methodReqT       string
		acceptEncodingT  string
		contentEncodingT string
		contentTypeT     string
		rawData          string
		wantStatusCode   int
	}{
		{
			nameT:            "Tx-gzip Rx-gzip",
			methodReqT:       http.MethodPost,
			acceptEncodingT:  "gzip",
			contentEncodingT: "gzip",
			contentTypeT:     "application/json",
			rawData:          "https://practicum.yandex.ru/",
			wantStatusCode:   http.StatusCreated,
		},
		{
			nameT:            "Tx-исходный Rx-gzip",
			methodReqT:       http.MethodPost,
			acceptEncodingT:  "",
			contentEncodingT: "gzip",
			contentTypeT:     "application/json",
			rawData:          "https://practicum.yandex.ru/",
			wantStatusCode:   http.StatusCreated,
		},
		{
			nameT:            "Tx-gzip Rx-исходный",
			methodReqT:       http.MethodPost,
			acceptEncodingT:  "gzip",
			contentEncodingT: "",
			contentTypeT:     "application/json",
			rawData:          "https://practicum.yandex.ru/",
			wantStatusCode:   http.StatusCreated,
		},
		{
			nameT:            "Tx-исходный Rx-исходный",
			methodReqT:       http.MethodPost,
			acceptEncodingT:  "",
			contentEncodingT: "",
			contentTypeT:     "application/json",
			rawData:          "https://practicum.yandex.ru/",
			wantStatusCode:   http.StatusCreated,
		},
	}

	for _, tt := range testData {
		t.Run(tt.nameT, func(t *testing.T) {

			txData := []byte(tt.rawData)
			var err error

			// Проверка необходимости компрессии
			if tt.contentEncodingT == "gzip" {
				txData, err = compress([]byte(tt.rawData))
				require.NoErrorf(t, err, "неожиданная ошибка при компрессии данных <%v>", err)
			}

			bodyReq := bytes.NewBuffer(txData)

			r := httptest.NewRequest(tt.methodReqT, srv.URL, bodyReq)
			r.RequestURI = ""
			r.Header.Set("Content-Encoding", tt.contentEncodingT)
			r.Header.Set("Accept-Encoding", tt.acceptEncodingT)
			r.Header.Set("Content-Type", tt.contentTypeT)

			resp, err := http.DefaultClient.Do(r)
			require.NoErrorf(t, err, "неожиданная ошибка запроса <%v>", err)
			require.Equalf(t, tt.wantStatusCode, resp.StatusCode, "ожидался код <%d>, а принят <%d>", tt.wantStatusCode, resp.StatusCode)

			defer func() {
				_ = resp.Body.Close()
			}()

			rxData, err := io.ReadAll(resp.Body)
			require.NoErrorf(t, err, "неожиданная ошибка при чтении тела ответа <%v>", err)

			// Проверка необходимости декомпрессии
			if tt.acceptEncodingT == "gzip" {
				rxData, err = decompress(rxData)
				require.NoErrorf(t, err, "неожиданная ошибка при декомпрессии данных <%v>", err)
			}

			// Проверка результата запроса
			shortRx := strings.TrimPrefix(string(rxData), "http://localhost"+shortLongHandler.BaseAddrShortURL)
			short, ok := shortLongHandler.List.ShorByLong[tt.rawData]
			require.Equalf(t, ok, true, "нет признака существования ключа в мапе")
			assert.Equalf(t, short, shortRx, "ожидалось <%s>, а приянто <%s>", short, shortRx)

		})
	}
}

func Test_Middleware_FAULT(t *testing.T) {

	shortLong := NewShortLongURL()
	shortLongHandler := &ShortLongT{
		List: shortLong,
	}

	handler := http.HandlerFunc(Middleware(shortLongHandler.ShortURLFromLong))

	srv := httptest.NewServer(handler)
	defer srv.Close()

	testData := []struct {
		nameT            string
		methodReqT       string
		acceptEncodingT  string
		contentEncodingT string
		contentTypeT     string
		rawData          string
		wantStatusCode   int
	}{
		{
			nameT:            "неподдерживаемый метод",
			methodReqT:       http.MethodGet,
			acceptEncodingT:  "gzip",
			contentEncodingT: "gzip",
			contentTypeT:     "application/json",
			rawData:          "https://practicum.yandex.ru/",
			wantStatusCode:   http.StatusBadRequest,
		},
		{
			nameT:            "неподдерживаея кодировка Rx",
			methodReqT:       http.MethodPost,
			acceptEncodingT:  "gzip",
			contentEncodingT: "AAA",
			contentTypeT:     "application/json",
			rawData:          "https://practicum.yandex.ru/",
			wantStatusCode:   http.StatusBadRequest,
		},
		{
			nameT:            "неподдерживаея кодировка Tx",
			methodReqT:       http.MethodPost,
			acceptEncodingT:  "AAA",
			contentEncodingT: "gzip",
			contentTypeT:     "application/json",
			rawData:          "https://practicum.yandex.ru/",
			wantStatusCode:   http.StatusBadRequest,
		},
		/*
			{
				nameT:            "неподдерживаемый тип контента",
				methodReqT:       http.MethodPost,
				acceptEncodingT:  "gzip",
				contentEncodingT: "gzip",
				contentTypeT:     "AAA",
				rawData:          "https://practicum.yandex.ru/",
				wantStatusCode:   http.StatusBadRequest,
			},
		*/
	}

	for _, tt := range testData {
		t.Run(tt.nameT, func(t *testing.T) {

			txData := []byte(tt.rawData)
			var err error

			// Проверка необходимости компрессии
			if tt.contentEncodingT == "gzip" {
				txData, err = compress([]byte(tt.rawData))
				require.NoErrorf(t, err, "неожиданная ошибка при компрессии данных <%v>", err)
			}

			bodyReq := bytes.NewBuffer(txData)

			r := httptest.NewRequest(tt.methodReqT, srv.URL, bodyReq)
			r.RequestURI = ""
			r.Header.Set("Content-Encoding", tt.contentEncodingT)
			r.Header.Set("Accept-Encoding", tt.acceptEncodingT)
			r.Header.Set("Content-Type", tt.contentTypeT)

			resp, err := http.DefaultClient.Do(r)
			require.NoErrorf(t, err, "ошибка запроса <%v>", err)
			defer func() {
				_ = resp.Body.Close()
			}()
			assert.Equalf(t, tt.wantStatusCode, resp.StatusCode, "ожидался код <%d>, а принят <%d>", tt.wantStatusCode, resp.StatusCode)

			_ = resp

		})
	}
}

func Test_LoadFileURL_SUCCESS(t *testing.T) {

	//flags := config.ParseFlags()

	shortLong := NewShortLongURL()
	shortLongHandler := &ShortLongT{
		List:             shortLong,
		BaseAddrShortURL: ":8080/",
		ServerAddr:       ":8080",
		FileStoragePath:  "storage.json",
	}

	testData := struct {
		originalURL1 string
		shortURL1    string
		originalURL2 string
		shortURL2    string
	}{
		originalURL1: "https://practicum.yandex.ru/",
		shortURL1:    "CyevslRg",
		originalURL2: "https://AAA.ru/",
		shortURL2:    "mZ0K-YfJ",
	}

	file, err := os.OpenFile(shortLongHandler.FileStoragePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	require.NoErrorf(t, err, "неожиданная ошибка при открытии файла <%v>", err)
	defer func() {
		_ = file.Close()
	}()

	// Заполнение файла
	file.WriteString("[\n")
	str := fmt.Sprintf(`	{"uuid":"%d","short_url":"%s","original_url":"%s"},`, 1, testData.shortURL1, testData.originalURL1)
	file.WriteString(str)
	file.WriteString("\n")
	str = fmt.Sprintf(`		{"uuid":"%d","short_url":"%s","original_url":"%s"}`, 2, testData.shortURL2, testData.originalURL2)
	file.WriteString(str)
	file.WriteString("\n")
	file.WriteString("]\n")

	// Чтение файла и проыерка результата
	err = shortLongHandler.LoadFileURL()
	require.NoErrorf(t, err, "ошибка чтения файла <%v>", err)

	v1, ok := shortLong.ShorByLong[testData.originalURL1]
	assert.Equalf(t, ok, true, "нет признака присутствия 1")

	v2, ok := shortLong.ShorByLong[testData.originalURL2]
	assert.Equalf(t, ok, true, "нет признака присутствия 2")

	assert.Equalf(t, testData.shortURL1, v1, "1: ожилось <%s>, а принято <%s>", testData.shortURL1, v1)
	assert.Equalf(t, testData.shortURL2, v2, "2: ожилось <%s>, а принято <%s>", testData.shortURL2, v2)

	// Удаление
	err = os.Remove(shortLongHandler.FileStoragePath)
	assert.NoErrorf(t, err, "неожиданная ошибка при удалении файла <%v>", err)
}

func Test_LoadFileURL_FAULT(t *testing.T) {

	testData := []struct {
		nameT         string
		mapsShortLong *ShortLongT
		wantError     string
	}{
		{
			nameT: "нет пути к файлу",
			mapsShortLong: &ShortLongT{
				List: &ShortLongURLT{
					ShorByLong:  map[string]string{},
					LongByShort: map[string]string{},
					Mu:          sync.RWMutex{},
				},
				BaseAddrShortURL: ":8080/",
				ServerAddr:       ":8080",
				FileStoragePath:  "",
			},
			wantError: "принят пустой путь к файлу хранения",
		},
		{
			nameT: "нет указателя на ShortByLong",
			mapsShortLong: &ShortLongT{
				List: &ShortLongURLT{
					ShorByLong:  nil,
					LongByShort: map[string]string{},
					Mu:          sync.RWMutex{},
				},
				BaseAddrShortURL: ":8080/",
				ServerAddr:       ":8080",
				FileStoragePath:  "test.json",
			},
			wantError: "нет указателя на ShortByLong",
		},
		{
			nameT: "нет указателя на LongByShort",
			mapsShortLong: &ShortLongT{
				List: &ShortLongURLT{
					ShorByLong:  map[string]string{},
					LongByShort: nil,
					Mu:          sync.RWMutex{},
				},
				BaseAddrShortURL: ":8080/",
				ServerAddr:       ":8080",
				FileStoragePath:  "test.json",
			},
			wantError: "нет указателя на LongByShort",
		},
	}

	for _, tt := range testData {
		t.Run(tt.nameT, func(t *testing.T) {

			err := tt.mapsShortLong.LoadFileURL()
			if err != nil {
				assert.Equalf(t, tt.wantError, err.Error(), "ожидалась ошибка <%s>, а принята <%s>", tt.wantError, err.Error())
			} else {
				assert.EqualError(t, err, tt.wantError)
			}
		})
	}
}

// Метрики

func Test_UpdateMetricByTypeAndName_SUCCESS(t *testing.T) {

	flags := config.ParseFlags()

	testMetrics := NewMetrics()
	m := NewMetricsStorage(testMetrics, flags)

	testsData := []struct {
		nameT          string
		methodT        string
		urlT           string
		wantStatusCode int
	}{
		{
			nameT:          "корректные данные 1",
			methodT:        http.MethodPost,
			urlT:           "http://localhost:8080/update/gauge/LastGC/1257894000000000000",
			wantStatusCode: http.StatusOK,
		},
		{
			nameT:          "корректные данные 2",
			methodT:        http.MethodPost,
			urlT:           "http://localhost:8080/update/counter/NumGC/42",
			wantStatusCode: http.StatusOK,
		},
	}
	for _, tt := range testsData {
		t.Run(tt.nameT, func(t *testing.T) {
			req := httptest.NewRequest(tt.methodT, tt.urlT, nil)
			res := httptest.NewRecorder()

			m.UpdateMetricByTypeAndName(res, req)
			resp := res.Result()
			defer func() {
				err := resp.Body.Close()
				assert.NoErrorf(t, err, "ошибка при закрытии потока {%v}", err)
			}()

			require.Equalf(t, tt.wantStatusCode, resp.StatusCode, "ожидался код {%d}, а принят {%d}", tt.wantStatusCode, resp.StatusCode)
		})
	}
}

func Test_UpdateMetricByTypeAndName_SUCCESS2(t *testing.T) {
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

func Test_MetricByJSON_SUCCESS(t *testing.T) {
	testMetrics := NewMetrics()

	metricsHandler := &MetricsHandlerT{
		Metrics: testMetrics,
	}
	testsData := []struct {
		nameT          string
		methodT        string
		urlT           string
		body           Metrics
		contentType    string
		valueMetric    float64
		wantStatusCode int
	}{
		{
			nameT:   "данные корректны",
			methodT: http.MethodPost,
			urlT:    "http://localhost:8080/update",
			body: Metrics{
				ID:    "LastGC",
				MType: "gauge",
				Value: new(float64),
			},
			contentType:    "application/json",
			valueMetric:    1744184459,
			wantStatusCode: http.StatusOK,
		},
	}
	for _, tt := range testsData {
		t.Run(tt.nameT, func(t *testing.T) {

			tt.body.Value = new(float64)
			*tt.body.Value = tt.valueMetric
			metricsHandler.Metrics.GaugeMetrics[tt.body.ID] = *tt.body.Value

			rawData, err := json.Marshal(tt.body)
			require.NoErrorf(t, err, "неожиданная ошибка сериализации <%v>", err)

			txData := bytes.NewBuffer(rawData)

			req := httptest.NewRequest(tt.methodT, tt.urlT, txData)
			res := httptest.NewRecorder()

			req.Header.Set("Content-Type", tt.contentType)

			metricsHandler.MetricByJSON(res, req)
			resp := res.Result()
			defer func() {
				err := resp.Body.Close()
				assert.NoErrorf(t, err, "ошибка при закрытии потока {%v}", err)
			}()

			require.Equalf(t, tt.wantStatusCode, resp.StatusCode, "ожидался код {%d}, а принят {%d}", tt.wantStatusCode, resp.StatusCode)

			var rxData Metrics

			err = json.NewDecoder(res.Body).Decode(&rxData)
			require.NoErrorf(t, err, "неожиданная ошибка десериализации <%v>", err)

			assert.Equalf(t, tt.body.ID, rxData.ID, "ожидался ID <%s> а принят <%s>", tt.body.ID, rxData.ID)
			assert.Equalf(t, tt.body.MType, rxData.MType, "ожидался MType <%s> а принят <%s>", tt.body.MType, rxData.MType)
			assert.Equalf(t, *tt.body.Value, *rxData.Value, "ожидался Value <%f> а принят <%f>", *tt.body.Value, *rxData.Value)

		})
	}
}

func Test_MetricByJSON_FAULT(t *testing.T) {
	testMetrics := NewMetrics()

	metricsHandler := &MetricsHandlerT{
		Metrics: testMetrics,
	}
	testsData := []struct {
		nameT          string
		methodT        string
		urlT           string
		metricsInit    Metrics
		body           Metrics
		contentType    string
		valueMetric    float64
		wantStatusCode int
	}{
		{
			nameT:   "неверный метод",
			methodT: http.MethodGet,
			urlT:    "http://localhost:8080/update",
			metricsInit: Metrics{
				ID:    "LastGC",
				MType: "gauge",
				Value: new(float64),
			},
			body: Metrics{
				ID:    "LastGC",
				MType: "gauge",
				Value: new(float64),
			},
			contentType:    "application/json",
			valueMetric:    1744184459,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			nameT:   "неверный контент",
			methodT: http.MethodPost,
			urlT:    "http://localhost:8080/update",
			metricsInit: Metrics{
				ID:    "LastGC",
				MType: "gauge",
				Value: new(float64),
			},
			body: Metrics{
				ID:    "LastGC",
				MType: "gauge",
				Value: new(float64),
			},
			contentType:    "application/AAA",
			valueMetric:    1744184459,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			nameT:   "нет содержимого в ID",
			methodT: http.MethodPost,
			urlT:    "http://localhost:8080/update",
			metricsInit: Metrics{
				ID:    "LastGC",
				MType: "gauge",
				Value: new(float64),
			},
			body: Metrics{
				ID:    "",
				MType: "gauge",
				Value: new(float64),
			},
			contentType:    "application/json",
			valueMetric:    1744184459,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			nameT:   "нет содержимого в MType",
			methodT: http.MethodPost,
			urlT:    "http://localhost:8080/update",
			metricsInit: Metrics{
				ID:    "LastGC",
				MType: "gauge",
				Value: new(float64),
			},
			body: Metrics{
				ID:    "LastGC",
				MType: "",
				Value: new(float64),
			},
			contentType:    "application/json",
			valueMetric:    1744184459,
			wantStatusCode: http.StatusBadRequest,
		},
	}
	for _, tt := range testsData {
		t.Run(tt.nameT, func(t *testing.T) {

			tt.metricsInit.Value = new(float64)
			*tt.metricsInit.Value = tt.valueMetric
			metricsHandler.Metrics.GaugeMetrics[tt.metricsInit.ID] = *tt.metricsInit.Value

			rawData, err := json.Marshal(tt.body)
			require.NoErrorf(t, err, "неожиданная ошибка сериализации <%v>", err)

			txData := bytes.NewBuffer(rawData)

			req := httptest.NewRequest(tt.methodT, tt.urlT, txData)
			res := httptest.NewRecorder()

			req.Header.Set("Content-Type", tt.contentType)

			metricsHandler.MetricByJSON(res, req)
			resp := res.Result()
			defer func() {
				err := resp.Body.Close()
				assert.NoErrorf(t, err, "ошибка при закрытии потока {%v}", err)
			}()

			assert.Equalf(t, tt.wantStatusCode, resp.StatusCode, "ожидася код <%d>, а принят <%d>", tt.wantStatusCode, resp.StatusCode)
		})
	}
}

// compress сжимает данные в формате gzip
func compress(data []byte) ([]byte, error) {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	defer w.Close()

	_, err := w.Write(data)
	if err != nil {
		return nil, fmt.Errorf("ошибка компрессии данных: %v", err)
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("ошибка закрытия gzip writer: %v", err)
	}

	return b.Bytes(), nil
}

// decompress распаковывает данные из формата gzip
func decompress(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("данные пустые")
	}

	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания gzip reader: %v", err)
	}
	defer r.Close()

	return io.ReadAll(r)
}
