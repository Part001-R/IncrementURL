package handler

import (
	"io"
	"net/http"
	"strings"
)

func HndlPOST(w http.ResponseWriter, r *http.Request) {
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

	strWant := "https://practicum.yandex.ru/"
	strRx := strings.Trim(string(rxData), "\"")

	if strRx != strWant {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	/*
		.
		.
	*/

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("http://localhost:8080/EwHXdJfB"))

}

func HndlGET(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/plain")

	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	rxData := r.PathValue("id")
	if len(rxData) == 0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if rxData != "EwHXdJfB" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	/*
		.
		.
	*/

	w.WriteHeader(http.StatusTemporaryRedirect)
	w.Write([]byte("Location: https://practicum.yandex.ru/"))
}
