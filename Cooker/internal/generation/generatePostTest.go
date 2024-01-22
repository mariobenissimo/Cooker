package generation

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"
	"go/token"
	"os"
	"strconv"
	"strings"

	"github.com/mariobenissimo/Cooker/internal/models"
)

func copyPayload(originalMap map[string]interface{}) map[string]interface{} {
	copyOfMap := make(map[string]interface{})

	for key, value := range originalMap {
		copyOfMap[key] = value
	}

	return copyOfMap
}
func GenerateTestEndpointPost(apiConfig models.APIConfig, nameFile string) {
	fmt.Println("Creazione del test per l'endpoint : " + apiConfig.Endpoint)

	// create an array of numParamters which contains nameField
	// create an array of numParamters which contains type of paramters
	// create an array for the first test with correct value
	// create array with ranges (optionally) and maxLengths (optionally)

	nameField, typeField, correctValue, ranges, maxLengths := extractNameAndType(apiConfig)

	endpoint := apiConfig.Endpoint
	// Check if Authentication is nil
	auth := false
	if apiConfig.Authentication != nil {
		auth = true
	}
	lim := false
	if apiConfig.Limiter != nil {
		lim = true
	}
	// inizilize variable if needed
	var minValue *int
	var MaxValue *int
	var randomString string
	var err error

	// create the right payload with the correct value
	correctPayload := make(map[string]interface{})

	for index, fieldName := range nameField {
		if typeField[index] == "int" {
			correctPayload[fieldName] = correctValue[index]
			// if type is int take minValue and MaxValue form ranges
			if minValue, MaxValue, err = sliptRange(*ranges[index]); err != nil {
				fmt.Println("Error: ", err)
			}
		}
		if typeField[index] == "string" {
			correctPayload[fieldName] = correctValue[index]
			// if type is string take random string with a > maxLenghts
			randomString = generateRandomString((*maxLengths[index]) + 1)
		}
	}

	// create a payload which i can modify, saved the correct
	payload := createGenericPayloadCode(correctPayload)
	if lim {
		// test rate limiter with post
		generateRateLimiterPost(auth, nameFile+"_rate_lim_post_test.go", "TestRateLimiterPost"+nameFile, "POST", endpoint, payload, apiConfig.Limiter.MaxRequests, apiConfig.Limiter.Seconds)
	}
	if auth {
		// se questo enpoint necessita di autenticazione anche se i parametri sono corretti dovrebbe dare come stato 401 non Autrizzato
		GenerateTestEndpointCorrectInt(!auth, endpoint, payload, HTTPSTATUS_UNAUTHORIZED, nameFile+"_gen_post_0_test.go", "TestPostUnAuthorized"+nameFile)
		// generate the file to get token
		GenerateTestTokenCode()
	}
	// generate the first endpoint with the correct payload
	GenerateTestEndpointCorrectInt(auth, endpoint, payload, HTTPSTATUS_CREATED, nameFile+"_gen_post_1_test.go", "TestPostCorrectValue"+nameFile)

	for index, value := range typeField {
		if value == "int" {

			genericPayload := copyPayload(correctPayload)
			genericPayload[nameField[index]] = *minValue //try to send a lower int value
			payload := createGenericPayloadCode(genericPayload)
			GenerateTestEndpointCorrectInt(auth, endpoint, payload, HTTPSTATUS_BADREQUEST, nameFile+"_gen_post_"+strconv.Itoa(index)+"2_test.go", "TestPostUpperIntValue"+strconv.Itoa(index)+nameFile)

			genericPayload[nameField[index]] = *MaxValue //try to send a upper Int value
			payload = createGenericPayloadCode(genericPayload)
			GenerateTestEndpointCorrectInt(auth, endpoint, payload, HTTPSTATUS_BADREQUEST, nameFile+"_gen_post_"+strconv.Itoa(index)+"3_test.go", "TestPostLowerIntValue"+strconv.Itoa(index)+nameFile)

			genericPayload[nameField[index]] = "aa" // try to send a string
			payload = createGenericPayloadCode(genericPayload)
			GenerateTestEndpointCorrectInt(auth, endpoint, payload, HTTPSTATUS_BADREQUEST, nameFile+"_gen_post_"+strconv.Itoa(index)+"4_test.go", "TestPostIncorrectIntValueString"+strconv.Itoa(index)+nameFile)

			genericPayload[nameField[index]] = 10.1 //try to send a a double
			payload = createGenericPayloadCode(genericPayload)
			GenerateTestEndpointCorrectInt(auth, endpoint, payload, HTTPSTATUS_BADREQUEST, nameFile+"_gen_post_"+strconv.Itoa(index)+"5_test.go", "TestPostIncorrectIntValueDouble"+strconv.Itoa(index)+nameFile)

			genericPayload[nameField[index]] = true //try to send a boolean
			payload = createGenericPayloadCode(genericPayload)
			GenerateTestEndpointCorrectInt(auth, endpoint, payload, HTTPSTATUS_BADREQUEST, nameFile+"_gen_post_"+strconv.Itoa(index)+"6_test.go", "TestPostIncorrectIntValueBoolean"+strconv.Itoa(index)+nameFile)
		}
		if value == "string" {

			genericPayload := copyPayload(correctPayload)
			genericPayload[nameField[index]] = randomString //try to send a > max len string
			payload = createGenericPayloadCode(genericPayload)
			GenerateTestEndpointCorrectInt(auth, endpoint, payload, HTTPSTATUS_BADREQUEST, nameFile+"_gen_post_"+strconv.Itoa(index)+"2_test.go", "TestPostUpperStringValue"+strconv.Itoa(index)+nameFile)

		}
	}
}
func GenerateTestEndpointCorrectInt(auth bool, endpoint string, payload ast.AssignStmt, statusCode string, nameFile string, nameTest string) {
	GenerateTestEndpointPostAuth(auth, endpoint, payload, statusCode, nameFile, nameTest)
}
func CreateJsonRequest() *ast.AssignStmt {
	return &ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.Ident{Name: "requestBody"},
			&ast.Ident{Name: "err"},
		},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "json"},
					Sel: &ast.Ident{Name: "Marshal"},
				},
				Args: []ast.Expr{&ast.Ident{Name: "requestPayload"}},
			},
		},
	}
}
func CreateNewHTTPRequestPayload(method, endpoint string) *ast.AssignStmt {
	return &ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.Ident{Name: "req"},
			&ast.Ident{Name: "err"},
		},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "http"},
					Sel: &ast.Ident{Name: "NewRequest"},
				},
				Args: []ast.Expr{
					&ast.BasicLit{Kind: token.STRING, Value: method},
					&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("\"%s\"", endpoint)},
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "bytes"},
							Sel: &ast.Ident{Name: "NewBuffer"},
						},
						Args: []ast.Expr{
							&ast.Ident{Name: "requestBody"},
						},
					},
				},
			},
		},
	}
}
func GenerateTestEndpointPostAuth(auth bool, endpoint string, payload ast.AssignStmt, statusCode string, nameFile string, nameTest string) {
	fset := token.NewFileSet()

	// Importa i pacchetti necessari
	importBytes := CreateImport(`"bytes"`)

	importJSON := CreateImport(`"encoding/json"`)

	importHTTP := CreateImport(`"net/http"`)

	importTesting := CreateImport(`"testing"`)

	importTest := CreateImport(`"github.com/stretchr/testify/assert"`)

	importDecl := &ast.GenDecl{
		Tok:   token.IMPORT,
		Specs: []ast.Spec{importTesting, importBytes, importJSON, importHTTP, importTest},
	}
	// Create requestBody, err := json.Marshal(requestPayload)
	marshalStmt := CreateJsonRequest()

	// Create the http.NewRequest statement
	newRequestStmt := CreateNewHTTPRequestPayload(`"POST"`, endpoint)

	// Create the http.DefaultClient.Do statement
	doRequestStmt := CreateHTTPDefaultClient()

	assertNoErrorStmt := CreateAssertError()

	tokenAssignment := CreateFunctionGetTestToken()

	firstHeader := CreateHeader(`"Content-Type"`, `"application/json"`)

	// Create the defer resp.Body.Close() statement
	deferStmt := CreateDeferBody()

	authHeaderAssignment := CreateHeaderToken()

	// Create the assert.Equal statement
	assertEqualStmt := CreateAssertEqualStatus(statusCode)

	// make discern
	var stmts = []ast.Stmt{}
	if auth {
		stmts = []ast.Stmt{
			&payload,
			marshalStmt,
			assertNoErrorStmt,
			newRequestStmt,
			assertNoErrorStmt,
			tokenAssignment,
			firstHeader,
			authHeaderAssignment,
			doRequestStmt,
			assertNoErrorStmt,
			deferStmt,
			assertEqualStmt,
		}
	} else {
		stmts = []ast.Stmt{
			&payload,
			marshalStmt,
			assertNoErrorStmt,
			newRequestStmt,
			assertNoErrorStmt,
			firstHeader,
			doRequestStmt,
			assertNoErrorStmt,
			deferStmt,
			assertEqualStmt,
		}
	}

	// Create a string builder to hold the generated source code
	var buf strings.Builder

	// Create sign of the function and the body
	funcDecl := CreateTest(nameTest, stmts)

	decls := []ast.Decl{importDecl, funcDecl}

	// Create a new file
	file := CreateFile(decls, "test")

	// Print code on buffer
	err := printer.Fprint(&buf, fset, file)
	if err != nil {
		fmt.Println("Error printing code:", err)
		return
	}

	// Format code
	formattedCode, err := format.Source([]byte(buf.String()))
	if err != nil {
		fmt.Println("Error formatting code:", err)
		return
	}

	// Create Folder if doesn't exist
	folderPath := "./testing"
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		err := os.MkdirAll(folderPath, 0755)
		if err != nil {
			fmt.Println("Error creating folder:", err)
			return
		}
	}

	// Specify the file path within the folder
	filePath := folderPath + "/" + nameFile

	// Crea os file and put code inside
	outputFile, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer outputFile.Close()

	_, err = outputFile.Write(formattedCode)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
	fmt.Println("Generated test code written to " + nameFile)
}

func createGenericPayloadCode(payload map[string]interface{}) ast.AssignStmt {
	return ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.Ident{Name: "requestPayload"},
		},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			generateMapLiteral(payload),
		},
	}
}

func generateMapLiteral(payload map[string]interface{}) *ast.CompositeLit {
	// structure map[string]interface{}
	mapType := &ast.MapType{
		Key:   &ast.Ident{Name: "string"},
		Value: &ast.InterfaceType{Methods: &ast.FieldList{}},
	}
	// create map[string]interface{}
	mapLiteral := &ast.CompositeLit{
		Type: mapType,
		Elts: []ast.Expr{},
	}
	// add correct element to map
	for key, value := range payload {
		mapLiteral.Elts = append(mapLiteral.Elts, &ast.KeyValueExpr{
			Key:   &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("\"%s\"", key)},
			Value: generateLiteral(value),
		})
	}
	return mapLiteral
}

func generateLiteral(value interface{}) ast.Expr {
	switch v := value.(type) {
	case int:
		return &ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", v)}
	case float64:
		return &ast.BasicLit{Kind: token.FLOAT, Value: fmt.Sprintf("%f", v)}
	case string:
		return &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("\"%s\"", v)}
	case bool:
		if v {
			return &ast.Ident{Name: "true"}
		}
		return &ast.Ident{Name: "false"}
	default:
		// Other types not handled
		return &ast.BasicLit{Kind: token.STRING, Value: "\"unsupported\""}
	}
}
