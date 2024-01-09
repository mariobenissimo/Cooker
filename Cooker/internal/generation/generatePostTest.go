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
	// with a correct int should return 200

	// first of all get the num of parameters suppose 2
	// numParamters := 2
	// create an array of numParamters which contains nameField
	nameField, typeField, correctValue, ranges, maxLengths := extractNameAndType(apiConfig)
	// create an array for the first test with correct value
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
	var minValue *int
	var MaxValue *int
	var randomString string
	var err error

	correctPayload := make(map[string]interface{})

	fmt.Println(correctValue)
	for index, fieldName := range nameField {
		if typeField[index] == "int" {
			correctPayload[fieldName] = correctValue[index]
			if minValue, MaxValue, err = sliptRange(*ranges[index]); err != nil {
				fmt.Println("Error: ", err)
			}
		}
		if typeField[index] == "string" {
			correctPayload[fieldName] = correctValue[index]
			randomString = generateRandomString((*maxLengths[index]) + 1)
		}
	}
	//print the generic payload wih correct value
	fmt.Println(correctPayload)
	// create first payload with correct value
	payload := createGenericPayloadCode(correctPayload)
	if lim {
		generateRateLimiterPost(auth, nameFile+"_rate_lim_post_test.go", "TestRateLimiterPost"+nameFile, "POST", endpoint, payload, apiConfig.Limiter.MaxRequests, apiConfig.Limiter.Seconds)
	}
	if auth {
		// se questo enpoint necessita di autenticazione anche se i parametri sono corretti dovrebbe dare come stato 401 non Autrizzato
		GenerateTestEndpointCorrectInt(!auth, endpoint, payload, "401", nameFile+"_gen_post_1_test.go", "TestPostUnAuthorized"+nameFile)
		// nel frattempo genero il codice che necessita per ottenere il toke
		GenerateTestTokenCode()
		GenerateTestEndpointCorrectInt(auth, endpoint, payload, "200", nameFile+"_gen_post_2_test.go", "TestPostCorrectValue"+nameFile)

	} else {
		GenerateTestEndpointCorrectInt(auth, endpoint, payload, "200", nameFile+"_gen_post_1_test.go", "TestPostCorrectValue"+nameFile)
	}

	for index, value := range typeField {
		if value == "int" {

			genericPayload := copyPayload(correctPayload)
			genericPayload[nameField[index]] = minValue //try to send a lower int value
			fmt.Println(genericPayload)
			payload := createGenericPayloadCode(genericPayload)
			GenerateTestEndpointCorrectInt(auth, endpoint, payload, "400", nameFile+"_gen_post_"+strconv.Itoa(index)+"2_test.go", "TestPostUpperIntValue"+strconv.Itoa(index)+nameFile)

			genericPayload[nameField[index]] = MaxValue //try to send a upper Int value
			fmt.Println(genericPayload)
			payload = createGenericPayloadCode(genericPayload)
			GenerateTestEndpointCorrectInt(auth, endpoint, payload, "400", nameFile+"_gen_post_"+strconv.Itoa(index)+"3_test.go", "TestPostLowerIntValue"+strconv.Itoa(index)+nameFile)

			genericPayload[nameField[index]] = "aa" // try to send a string
			fmt.Println(genericPayload)
			payload = createGenericPayloadCode(genericPayload)
			GenerateTestEndpointCorrectInt(auth, endpoint, payload, "400", nameFile+"_gen_post_"+strconv.Itoa(index)+"4_test.go", "TestPostIncorrectIntValueString"+strconv.Itoa(index)+nameFile)

			genericPayload[nameField[index]] = 10.1 //try to send a a double
			fmt.Println(genericPayload)
			payload = createGenericPayloadCode(genericPayload)
			GenerateTestEndpointCorrectInt(auth, endpoint, payload, "400", nameFile+"_gen_post_"+strconv.Itoa(index)+"5_test.go", "TestPostIncorrectIntValueDouble"+strconv.Itoa(index)+nameFile)

			genericPayload[nameField[index]] = true //try to send a boolean
			fmt.Println(genericPayload)
			payload = createGenericPayloadCode(genericPayload)
			GenerateTestEndpointCorrectInt(auth, endpoint, payload, "400", nameFile+"_gen_post_"+strconv.Itoa(index)+"6_test.go", "TestPostIncorrectIntValueBoolean"+strconv.Itoa(index)+nameFile)
		}
		if value == "string" {

			genericPayload := copyPayload(correctPayload)
			genericPayload[nameField[index]] = randomString //try to send a > max len string
			payload = createGenericPayloadCode(genericPayload)
			GenerateTestEndpointCorrectInt(auth, endpoint, payload, "400", nameFile+"_gen_post_"+strconv.Itoa(index)+"2_test.go", "TestPostUpperStringValue"+strconv.Itoa(index)+nameFile)

		}
	}
}
func GenerateTestEndpointCorrectInt(auth bool, endpoint string, payload ast.AssignStmt, statusCode string, nameFile string, nameTest string) {
	GenerateTestEndpointPostAuth(auth, endpoint, payload, statusCode, nameFile, nameTest)
}

func GenerateTestEndpointPostAuth(auth bool, endpoint string, payload ast.AssignStmt, statusCode string, nameFile string, nameTest string) {
	fset := token.NewFileSet()

	// Importa i pacchetti necessari
	importBytes := &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: `"bytes"`,
		},
	}

	importJSON := &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: `"encoding/json"`,
		},
	}

	importHTTP := &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: `"net/http"`,
		},
	}

	importTesting := &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: `"testing"`,
		},
	}
	importTest := &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: `"github.com/stretchr/testify/assert"`,
		},
	}

	importDecl := &ast.GenDecl{
		Tok:   token.IMPORT,
		Specs: []ast.Spec{importTesting, importBytes, importJSON, importHTTP, importTest},
	}
	// Create the request payload
	marshalStmt := &ast.AssignStmt{
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

	// Create the http.NewRequest statement
	newRequestStmt := &ast.AssignStmt{
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
					&ast.BasicLit{Kind: token.STRING, Value: `"POST"`},
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

	// Create the http.DefaultClient.Do statement
	doRequestStmt := &ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.Ident{Name: "resp"},
			&ast.Ident{Name: "err"},
		},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.SelectorExpr{X: &ast.Ident{Name: "http"}, Sel: &ast.Ident{Name: "DefaultClient"}},
					Sel: &ast.Ident{Name: "Do"},
				},
				Args: []ast.Expr{
					&ast.Ident{Name: "req"},
				},
			},
		},
	}

	assertNoErrorStmt := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "assert"},
				Sel: &ast.Ident{Name: "NoError"},
			},
			Args: []ast.Expr{&ast.Ident{Name: "t"}, &ast.Ident{Name: "err"}},
		},
	}

	tokenAssignment := &ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.Ident{Name: "token"},
		},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.Ident{Name: "GetTestToken"},
			},
		},
	}

	firstHeader := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.SelectorExpr{X: &ast.Ident{Name: "req"}, Sel: &ast.Ident{Name: "Header"}},
				Sel: &ast.Ident{Name: "Set"},
			},
			Args: []ast.Expr{
				&ast.BasicLit{Kind: token.STRING, Value: `"Content-Type"`},
				&ast.BasicLit{Kind: token.STRING, Value: `"application/json"`},
			},
		},
	}

	// Create the defer resp.Body.Close() statement
	deferStmt := &ast.DeferStmt{
		Call: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.SelectorExpr{X: &ast.Ident{Name: "resp"}, Sel: &ast.Ident{Name: "Body"}},
				Sel: &ast.Ident{Name: "Close"},
			},
		},
	}

	authHeaderAssignment := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.SelectorExpr{X: &ast.Ident{Name: "req"}, Sel: &ast.Ident{Name: "Header"}},
				Sel: &ast.Ident{Name: "Set"},
			},
			Args: []ast.Expr{
				&ast.BasicLit{Kind: token.STRING, Value: `"Authorization"`},
				&ast.BinaryExpr{
					X:  &ast.BasicLit{Kind: token.STRING, Value: `"Bearer "`},
					Op: token.ADD,
					Y:  &ast.Ident{Name: "token"},
				},
			},
		},
	}

	// Create the assert.Equal statement
	assertEqualStmt := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "assert"},
				Sel: &ast.Ident{Name: "Equal"},
			},
			Args: []ast.Expr{
				&ast.Ident{Name: "t"},
				&ast.BasicLit{Kind: token.STRING, Value: statusCode},
				&ast.SelectorExpr{
					X:   &ast.Ident{Name: "resp"},
					Sel: &ast.Ident{Name: "StatusCode"},
				},
			},
		},
	}

	// make discern
	stmts := []ast.Stmt{}
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

	// Create a new function
	funcDecl := &ast.FuncDecl{
		Name: &ast.Ident{Name: nameTest},
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					{Names: []*ast.Ident{{Name: "t"}}, Type: &ast.Ident{Name: "*testing.T"}},
				},
			},
			Results: &ast.FieldList{},
		},
		Body: &ast.BlockStmt{
			List: stmts,
		},
	}
	decls := []ast.Decl{importDecl, funcDecl}

	// Create a new file
	file := &ast.File{
		Name:  &ast.Ident{Name: "main"},
		Decls: decls,
	}

	// Aggiungi le importazioni e la funzione di test alla lista di dichiarazioni

	// Stampa il codice sorgente generato nel buffer
	err := printer.Fprint(&buf, fset, file)
	if err != nil {
		fmt.Println("Error printing code:", err)
		return
	}

	// Formatta il codice sorgente nel buffer
	formattedCode, err := format.Source([]byte(buf.String()))
	if err != nil {
		fmt.Println("Error formatting code:", err)
		return
	}

	// Crea la cartella tetsing se non esiste
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

	// Crea un file e scrivi il codice generato al suo interno
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

	a := ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.Ident{Name: "requestPayload"},
		},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			generateMapLiteral(payload),
		},
	}
	return a
}

func generateMapLiteral(payload map[string]interface{}) *ast.CompositeLit {
	// Creazione di una struttura di tipo map[string]interface{}
	mapType := &ast.MapType{
		Key:   &ast.Ident{Name: "string"},
		Value: &ast.InterfaceType{Methods: &ast.FieldList{}},
	}

	// Creazione di un valore di tipo map[string]interface{}
	mapLiteral := &ast.CompositeLit{
		Type: mapType,
		Elts: []ast.Expr{},
	}

	// Aggiunta degli elementi alla mappa
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
		// Other types not handled in this example
		return &ast.BasicLit{Kind: token.STRING, Value: "\"unsupported\""}
	}
}
