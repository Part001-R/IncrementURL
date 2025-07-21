package handler

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/Part001-R/IncrementURL/internal/service"
)

type ShortLongT struct {
	List             *service.ShortLongURL
	BaseAddrShortURL string
	ServerAddr       string
	mu               sync.RWMutex
}

func (sl *ShortLongT) ShortURLFromLong(w http.ResponseWriter, r *http.Request) {

	sl.mu.RLock()
	defer sl.mu.RUnlock()

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

	short, err := fillListsLongShort(sl.List.ShorByLong, sl.List.LongByShort, rxLongURL)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	strResult := "http://localhost" + sl.BaseAddrShortURL + short

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(strResult))

}

func (sl *ShortLongT) LongURLFromShort(w http.ResponseWriter, r *http.Request) {

	sl.mu.RLock()
	defer sl.mu.RUnlock()

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

	long, ok := sl.List.LongByShort[short]
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", long)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// Заполнение мап соответствий
func fillListsLongShort(sByL map[string]string, lByS map[string]string, longURL string) (string, error) {

	/*
		if v, ok := sByL[longURL]; ok {
			return v, nil
		}
	*/

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
