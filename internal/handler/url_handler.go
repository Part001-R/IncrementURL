package handler

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Part001-R/IncrementURL/internal/config/config"
	gz "github.com/Part001-R/IncrementURL/internal/service/gzip"
	"github.com/Part001-R/IncrementURL/internal/service/logger"
	"go.uber.org/zap"
)

type ShortLongURLT struct {
	ShorByLong  map[string]string
	LongByShort map[string]string
	Mu          sync.RWMutex
}

type ShortLongT struct {
	List             *ShortLongURLT
	BaseAddrShortURL string
	ServerAddr       string
	FileStoragePath  string
}

type EventURLT struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type rxLongURLT struct {
	URL string `json:"url"`
}

type txShortURLT struct {
	Result string `json:"result"`
}

type ShortLongI interface {
	ShortURLFromLong(w http.ResponseWriter, r *http.Request)
	LongURLFromShort(w http.ResponseWriter, r *http.Request)
	ShortURLFromLongJSON(w http.ResponseWriter, r *http.Request)
	LoadFileURL() error
}

func NewShortLongURL() *ShortLongURLT {
	return &ShortLongURLT{
		ShorByLong:  make(map[string]string),
		LongByShort: make(map[string]string),
		Mu:          sync.RWMutex{},
	}
}

func NewShortLongStorage(storage *ShortLongURLT, Fl config.FlagsT) ShortLongI {
	return &ShortLongT{
		List:             storage,
		BaseAddrShortURL: Fl.BaseAddrShortURL,
		ServerAddr:       Fl.ServerAddr,
		FileStoragePath:  Fl.FileStoragePath,
	}
}

func (sl *ShortLongT) ShortURLFromLong(w http.ResponseWriter, r *http.Request) {

	sl.List.Mu.RLock()
	defer sl.List.Mu.RUnlock()

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

	rxLongURL := string(rxData)

	short, err := fillListShortByLong(sl.List.ShorByLong, sl.List.LongByShort, rxLongURL)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	strResult := "http://localhost" + sl.BaseAddrShortURL + short

	err = storageFileURL(sl.FileStoragePath, sl.List.ShorByLong)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(strResult))

}

func (sl *ShortLongT) LongURLFromShort(w http.ResponseWriter, r *http.Request) {

	sl.List.Mu.RLock()
	defer sl.List.Mu.RUnlock()

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

	long, ok := sl.List.LongByShort[short]
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	long = strings.Trim(long, "\"")

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Location", long)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (sl *ShortLongT) ShortURLFromLongJSON(w http.ResponseWriter, r *http.Request) {
	sl.List.Mu.RLock()
	defer sl.List.Mu.RUnlock()

	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if r.Header.Get("Content-Type") != `application/json` {
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

	var rxJSON = rxLongURLT{}
	err = json.Unmarshal(rxData, &rxJSON)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if rxJSON.URL == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	short, err := fillListShortByLong(sl.List.ShorByLong, sl.List.LongByShort, rxJSON.URL)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	strResult := "http://localhost" + sl.BaseAddrShortURL + short
	var txJSON = txShortURLT{
		Result: strResult,
	}
	txData, err := json.Marshal(txJSON)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/jsom")
	w.WriteHeader(http.StatusCreated)
	w.Write(txData)

}

func (sl *ShortLongT) LoadFileURL() error {
	sl.List.Mu.RLock()
	defer sl.List.Mu.RUnlock()

	// Проверка
	if sl.FileStoragePath == "" {
		return errors.New("принят пустой путь к файлу хранения")
	}
	if sl.List.ShorByLong == nil {
		return errors.New("нет указателя на ShortByLong")
	}
	if sl.List.LongByShort == nil {
		return errors.New("нет указателя на LongByShort")
	}

	// Файл
	file, err := os.OpenFile(sl.FileStoragePath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("ошибка открытия файла <%s>: %v", sl.FileStoragePath, err)
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

	// Мапы
	var events []EventURLT
	if err := json.Unmarshal(data, &events); err != nil {
		return fmt.Errorf("ошибка Unmarshal: %v", err)
	}

	for _, ev := range events {
		sl.List.ShorByLong[ev.OriginalURL] = ev.ShortURL
		sl.List.LongByShort[ev.ShortURL] = ev.OriginalURL
	}
	return nil
}

func Middleware(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		// Проверка поддержки типа контента
		/*
			contentType := r.Header.Get("Content-Type")

			switch contentType {
			case "application/json", "text/html", "text/plain":

			default:
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
		*/

		// Проверка поддерживает ли сервер запрашиваемую клиентом кодировку
		acceptEncoding := r.Header.Get("Accept-Encoding")
		found := false

		if acceptEncoding != "" {
			encodings := strings.Split(acceptEncoding, ",")
			for _, v := range encodings {
				encodingType := strings.TrimSpace(v)

				switch encodingType {
				case "gzip":
					cw := gz.NewCompressWriter(w)
					ow = cw
					defer func() {
						if err := cw.Close(); err != nil {
							logger.Log.Error("Ошибка при закрытии cw", zap.Error(err))
						}
					}()
					found = true
				default:
				}
			}

			if !found {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
		}

		// Проверка, как клиент закодировал переданные данные
		contentEncoding := r.Header.Get("Content-Encoding")
		found = false

		if contentEncoding != "" {
			encodings := strings.Split(contentEncoding, ",")
			for _, v := range encodings {
				encodingType := strings.TrimSpace(v)

				switch encodingType {
				case "gzip":
					cr, err := gz.NewCompressReader(r.Body)
					if err != nil {
						http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
						return
					}
					defer func() {
						if err := cr.Close(); err != nil {
							logger.Log.Error("Ошибка при закрытии cr", zap.Error(err))
						}
					}()
					defer func() {
						if err := r.Body.Close(); err != nil {
							logger.Log.Error("Ошибка при закрытии r.Body", zap.Error(err))
						}
					}()

					r.Body = cr
					found = true

				default:
				}
			}

			if !found {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
		}

		// Запуск обработчика
		timeStart := time.Now()
		h(ow, r)
		duration := time.Since(timeStart)

		// Вывод в лог сводной информации по запросу
		logger.Log.Info("принят HTTP запрос",
			zap.String("URI", r.RequestURI),
			zap.String("метод", r.Method),
			zap.Duration("время выполнения запроса", duration),
		)
	})
}

// Заполнение мап соответствий
func fillListShortByLong(sByL, lByS map[string]string, longURL string) (string, error) {

	var short string
	var err error

	for {
		short, err = generateCode(8)
		if err != nil {
			return "", fmt.Errorf("ошибка при генерации короткой ссылки: {%w}", err)
		}

		if _, ok := lByS[short]; ok {
			continue
		}
		sByL[longURL] = short

		// Проверка, что в мапе есть уже значение
		// удаление, если есть
		for k, v := range lByS {
			if v == longURL {
				delete(lByS, k)
			}
		}

		lByS[short] = longURL
		break
	}

	return short, nil
}

// Генерация случайных символов заданной длинны
func generateCode(n int) (string, error) {
	b := make([]byte, n)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b)[:n], nil
}

// Обновление содержимого файла хранения данных
func storageFileURL(filename string, mapShortByLong map[string]string) error {

	if filename == "" {
		return errors.New("принят пустой путь к файлу хранения")
	}
	if mapShortByLong == nil {
		return errors.New("нет указателя на мапу")
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("ошибка <%v> открытия файла <%s>", err, filename)
	}
	defer func() {
		_ = file.Close()
	}()

	// Вывод с построением строк
	file.WriteString("[\n")

	numb := 1
	size := len(mapShortByLong)
	for k, v := range mapShortByLong {

		baseURL := strings.Trim(k, "\"")

		str := fmt.Sprintf(`	{"uuid":"%s","short_url":"%s","original_url":"%s"}`, fmt.Sprintf("%d", numb), v, baseURL)
		file.WriteString(str)
		if numb < size {
			file.WriteString(",\n")
		}
		numb++
	}
	file.WriteString("\n")
	file.WriteString("]\n")

	/*
		// Реализация с выводом в одну строку

		var events []EventURLT
		numb := 1
		for k, v := range mapShortByLong {

			var event EventURLT
			event.UUID = numb
			event.OriginalURL = k
			event.ShortURL = v

			events = append(events, event)
			numb++
		}

		encoder := json.NewEncoder(file)

		if err := encoder.Encode(events); err != nil {
			return err
		}
	*/

	return nil
}
