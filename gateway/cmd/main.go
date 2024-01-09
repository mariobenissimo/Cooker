package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"gitub.com/mariobenissimo/apiGateway/internal/handlers"
	"gitub.com/mariobenissimo/apiGateway/internal/middleware"
)

func main() {

	r := mux.NewRouter()

	r.HandleFunc("/user", handlers.UserHandler).Methods("GET")

	r.HandleFunc("/user/{id}", handlers.UserHandler).Methods("GET")
	// create a new user
	s := r.PathPrefix("/auth").Subrouter()
	s.HandleFunc("/user", handlers.UserHandler).Methods("POST")

	s.Use(middleware.AuthMiddleware)
	log.Fatal(http.ListenAndServe(":8000", r))
}
