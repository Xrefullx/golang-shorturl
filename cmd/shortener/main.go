package main

import (
	"github.com/Xrefullx/golang-shorturl/internal/handlers"
	"github.com/Xrefullx/golang-shorturl/internal/storage"
	"log"
	"net/http"
)

func main() {
	handlers := handlers.Handler{DB: storage.NewStorage()}
	http.HandleFunc("/", handlers.CheckRequestHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
