package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	gen "github.com/mariobenissimo/Cooker/internal/generation"
	"github.com/mariobenissimo/Cooker/internal/models"
)

func Start(c models.Cooker) {
	r := mux.NewRouter()

	r.HandleFunc("/json", CreateCookerFolder).Methods("POST")

	// if the frontend mode is active view on the / and /json to get json saved
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("../public/views"))))

	log.Fatal(http.ListenAndServe(c.Port, r))
}
func CreateCookerFolder(w http.ResponseWriter, r *http.Request) {
	// goal to get the json from frontend annd create folder cooker with json
	var requestData []models.APIConfig
	// // Decode JSON from the request body
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&requestData)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Print the received JSON
	fmt.Printf("Received JSON: %+v\n", requestData)
	// for every json received create the file
	for i, item := range requestData {
		saveJSONToFile(item, i)
	}
	// You can send a response back if needed
	gen.Cook("../cooker")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "JSON received successfully"}`))
}
func saveJSONToFile(data models.APIConfig, index int) {
	// Create a new folder named "json_files" if it doesn't exist
	folderPath := "../cooker"
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		os.Mkdir(folderPath, 0755)
	}

	// Generate a unique filename based on the current timestamp
	filename := fmt.Sprintf("%d_endpoint.json", index)

	// Join the folder path and filename
	filePath := filepath.Join(folderPath, filename)

	// Marshal the JSON data to a byte slice
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	// Write the JSON data to the file
	err = ioutil.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	fmt.Printf("JSON saved to: %s\n", filePath)
}
