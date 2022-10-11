package main

import (
	"context"
	"net/http"
	"time"

	"notebook/internals/handlers"
	"notebook/internals/storage"

	"github.com/gorilla/mux"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	connStr := "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable"
	if err := storage.DBInit(ctx, connStr); err != nil {
		panic(err)
	}
	ns, err := storage.New(ctx, connStr)
	if err != nil {
		panic(err)
	}
	defer ns.Close()
	h := handlers.NewHandler(ns)

	ih, err := handlers.NewImageHandler("./static")
	if err != nil {
		panic(err)
	}

	r := mux.NewRouter()

	n := r.PathPrefix("/api/v1/notebook").Subrouter() // n for notes
	n.HandleFunc("/", h.ListRange).Methods("GET")
	n.HandleFunc("/", h.Add).Methods("POST")
	n.HandleFunc("/{id}/", h.List).Methods("GET")
	n.HandleFunc("/{id}/", h.Update).Methods("POST")
	n.HandleFunc("/{id}/", h.Delete).Methods("DELETE")

	prefix := "/api/v1/images/"
	stripper := func(next http.Handler) http.Handler {
		return http.StripPrefix(prefix, next)
	}
	i := r.PathPrefix(prefix).Subrouter() // i for images
	i.Use(mux.MiddlewareFunc(stripper))
	i.HandleFunc("/", ih.Add).Methods("POST")
	i.HandleFunc("/{id}/", ih.Get).Methods("GET")
	i.HandleFunc("/{id}/", ih.Delete).Methods("DELETE")

	http.ListenAndServe(":8080", r)
}
