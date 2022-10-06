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
	h := handlers.New(ns)

	mux := mux.NewRouter().PathPrefix("/api/v1/notebook").Subrouter()
	mux.HandleFunc("/", h.ListRange).Methods("GET")
	mux.HandleFunc("/", h.Add).Methods("POST")
	mux.HandleFunc("/{id}/", h.List).Methods("GET")
	mux.HandleFunc("/{id}/", h.Update).Methods("POST")
	mux.HandleFunc("/{id}/", h.Delete).Methods("DELETE")

	http.ListenAndServe(":8080", mux)
}
