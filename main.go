package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	log.SetFlags(log.Lshortfile | log.Ltime)
	r := chi.NewRouter()

	log.Println("started at :8080")
	log.Println(http.ListenAndServe(":8080", r))
}
