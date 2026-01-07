package main

import (
	"log"
	"microblog/internal/handlers"
	"microblog/internal/service"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	svc := service.NewService()
	h := handlers.NewHandlers(svc)

	r := mux.NewRouter()
	h.SetupRoutes(r)

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}