package main

import (
	"net/http"

	"github.com/Part001-R/IncrementURL/internal/handler"
	"github.com/go-chi/chi/v5"
)

func main() {

	cr := chi.NewRouter()

	cr.Post("/", handler.HndlPOST)
	cr.Get("/{id}", handler.HndlGET)

	http.ListenAndServe(":8080", cr)
}
