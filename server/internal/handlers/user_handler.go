package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/mariobenissimo/AutoTest/internal/models"
)

// check if a string is longer tha a maxLenght
func isLongerThan(value string, maxLength int) bool {
	return len(value) > maxLength
}

// handler to create a new user
func AddUser(w http.ResponseWriter, r *http.Request) {

	var requestData struct {
		Età  int    `json:"età"`
		Nome string `json:"nome"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}
	if requestData.Età < 0 || requestData.Età > 100 {
		http.Error(w, "Value must be between 0 and 100", http.StatusBadRequest)
		return
	}
	if isLongerThan(requestData.Nome, 50) {
		http.Error(w, "No valid String", http.StatusBadRequest)
		return
	}
	models.AddUser(models.User{ID: uuid.New(), Età: requestData.Età, Nome: requestData.Nome})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

// handler to get a user by id
func GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	uuidValue, err := uuid.Parse(id)
	if err != nil {
		fmt.Println("Error converting string to UUID:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user := models.SearchUserByID(uuidValue)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

// handler to get all users
func GetAllUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.UsersDatabase)
}
