package generation

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/mariobenissimo/Cooker/internal/models"
)

func Cook(directoryPath string) {
	// Read and process all JSON files in the specified directory
	err := filepath.Walk(directoryPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Process only .json files
		if filepath.Ext(path) == ".json" {
			// Read the content of the JSON file
			nameFile := strings.TrimSuffix(filepath.Base(path), ".json")

			jsonData, err := ioutil.ReadFile(path)
			if err != nil {
				fmt.Println("Error reading JSON file:", err)
				return nil
			}

			// Create an instance of YourStruct to unmarshal the JSON data into
			var apiConfig models.APIConfig

			// Unmarshal the JSON data into the struct
			err = json.Unmarshal(jsonData, &apiConfig)
			if err != nil {
				fmt.Println("Error unmarshaling JSON in file", path, ":", err)
				return nil
			}

			// Print the populated struct
			fmt.Printf("File: %s\n", path)
			err = checkApiConfig(apiConfig)
			if err != nil {
				fmt.Println("Error checking correcting json", err)
				return nil
			}
			createTest(apiConfig, nameFile)
		}

		return nil
	})

	if err != nil {
		fmt.Println("Error walking the path:", err)
	}
}

// Lista di metodi endpoint
var methods = []string{"GET", "POST", "PUT", "DELETE"}

// Controlla se un metodo fa parte della lista di metodi endpoint
func isAllowMethods(method string) error {
	// Confronta il metodo in modo case-insensitive
	method = strings.ToUpper(method)
	for _, m := range methods {
		if m == method {
			return nil
		}
	}
	return ErrMethodNotAllow
}

var ErrMethodNotAllow = errors.New("method not allow")

var ErrAutenticationMethodNotAllow = errors.New("autentication method not allow")

var autenticationMethods = []string{"JWT"}

func isAllowAutenticationMethods(method string) error {
	// Confronta il metodo in modo case-insensitive
	method = strings.ToUpper(method)
	for _, m := range autenticationMethods {
		if m == method {
			return nil
		}
	}
	return ErrAutenticationMethodNotAllow
}

var ErrTypeNotAllow = errors.New("type not allow")

var types = []string{"STRING", "INT", "UUID"}

func isAllowType(typ string) error {
	// Confronta il metodo in modo case-insensitive
	typ = strings.ToUpper(typ)
	for _, m := range types {
		if m == typ {
			return nil
		}
	}
	return ErrTypeNotAllow
}

var ErrNoMaxLengthProvided = errors.New("no max length for string provided")
var ErrNoRangeProvided = errors.New("no range for int provided")
var ErrNegativeMaxLengthProvided = errors.New("max length negative provided")
var ErrInvalidRangeProvided = errors.New("max length negative provided")
var ErrNoExpectationLengthProvided = errors.New("no expectation length provided")

func isValidRangePattern(pattern *string) error {
	// Check if the pattern matches the format "number - number" or "number-number"
	re := regexp.MustCompile(`^\s*(\d+)\s*-\s*(\d+)\s*$`)
	matches := re.FindStringSubmatch(*pattern)
	if len(matches) != 3 {
		return ErrInvalidRangeProvided
	}

	// Check if both parts are valid integers
	_, err1 := strconv.Atoi(matches[1])
	if err1 != nil {
		return ErrInvalidRangeProvided
	}
	_, err2 := strconv.Atoi(matches[2])
	if err2 != nil {
		return ErrInvalidRangeProvided
	}
	return nil
}

func isCorrectFieldProvided(parameter models.Parameter, method string) error {
	if parameter.Type == "string" {
		//check for maxLength provided
		if parameter.MaxLength == nil {
			return ErrNoMaxLengthProvided
		}
		// check if is a positive number
		if *parameter.MaxLength <= 0 {
			return ErrNegativeMaxLengthProvided
		}
	}
	if parameter.Type == "int" {
		//check for range provided
		if parameter.Range == nil {
			return ErrNoRangeProvided
		}
		//check if pattern "number-number" provided
		if err := isValidRangePattern(parameter.Range); err != nil {
			return err
		}
	}
	return nil
}
func checkApiConfig(apiConfig models.APIConfig) error {
	// eseguo la scansione della struttura per vedere se la struttura json è conferme
	// controllo se il metodo fornito fa parte dei metodi accetati
	if err := isAllowMethods(apiConfig.Method); err != nil {
		return err
	}
	if apiConfig.Authentication != nil {
		if err := isAllowAutenticationMethods(apiConfig.Authentication.Method); err != nil {
			return err
		}
	}
	if apiConfig.Method == "get" {
		if apiConfig.ExpectationLength == nil {
			return ErrNoExpectationLengthProvided
		}
	}
	for _, parameter := range apiConfig.Parameters {
		if err := isAllowType(parameter.Type); err != nil {
			return err
		}
		// check for correct value provided
		if err := isCorrectFieldProvided(parameter, apiConfig.Method); err != nil {
			return err
		}
	}
	return nil
}
func createTest(apiConfig models.APIConfig, nameFile string) {
	if apiConfig.Method == "get" {
		GenerateTestEndpointGet(apiConfig, nameFile)
	} else if apiConfig.Method == "post" {
		GenerateTestEndpointPost(apiConfig, nameFile)
	}
}
