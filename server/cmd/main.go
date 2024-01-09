package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mariobenissimo/AutoTest/internal/handlers"
	"github.com/mariobenissimo/AutoTest/internal/models"
)

func main() {

	models.Setup()
	// before inizialize mux create two dummy user
	r := mux.NewRouter()

	// get alla users from db
	r.HandleFunc("/user", handlers.GetAllUser).Methods("GET")
	// // get user with a specific uiid
	r.HandleFunc("/user/{id}", handlers.GetUser).Methods("GET")
	// create a new user
	r.HandleFunc("/auth/user", handlers.AddUser).Methods("POST")

	log.Fatal(http.ListenAndServe(":8088", r))
}
