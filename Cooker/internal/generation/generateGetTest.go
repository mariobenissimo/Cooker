package generation

import (
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"
	"go/token"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mariobenissimo/Cooker/internal/models"
)

func replaceField(correctEndpoint, endpoint, newField string, index int) string {
	endpointWhitoutPrefix := strings.TrimPrefix(correctEndpoint, endpoint+"/")

	// Split the URL by "/"
	parts := strings.Split(endpointWhitoutPrefix, "/")

	parts[index] = newField
	// Join the modified parts back into a URL
	newURL := strings.Join(parts, "/")

	return endpoint + "/" + newURL
}
func extractNameAndType(apiConfig models.APIConfig) ([]string, []string, []interface{}, [](*string), [](*int)) {
	var names []string
	var types []string
	var correctValue []interface{}
	var ranges [](*string)
	var maxLengths [](*int)
	for _, param := range apiConfig.Parameters {
		names = append(names, param.Name)
		types = append(types, param.Type)
		correctValue = append(correctValue, param.CorrectValue)
		ranges = append(ranges, param.Range)
		maxLengths = append(maxLengths, param.MaxLength)
	}
	return names, types, correctValue, ranges, maxLengths
}
func generateRandomString(length int) string {
	// Define the character set for the random string
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Generate the random string
	randomString := make([]byte, length)
	for i := range randomString {
		randomString[i] = charset[rand.Intn(len(charset))]
	}

	return string(randomString)
}

var ErrParsingRange = errors.New("error parsing number")

func sliptRange(str string) (*int, *int, error) {
	// Split the string by "-"
	parts := strings.Split(str, "-")

	if len(parts) == 2 {
		// Convert the parts to integers
		num1, err1 := strconv.Atoi(parts[0])
		num2, err2 := strconv.Atoi(parts[1])

		// Check for conversion errors
		if err1 == nil && err2 == nil {
			num1 -= 1
			num2 += 1
			return &num1, &num2, nil
		} else {
			return nil, nil, ErrParsingRange
		}
	} else {
		return nil, nil, ErrParsingRange
	}
}
func GenerateTestEndpointGet(apiConfig models.APIConfig, nameFile string) {
	fmt.Println("Creazione del test per l'endpoint : " + apiConfig.Endpoint)
	// with a correct int should return 200

	// first of all get the num of parameters suppose 2
	// numParamters := 2
	// create an array of numParamters which contains nameField
	//nameField := []string{"id"}
	// create an array of numParamters which contains type of paramters
	//typeField := []string{"uuid"}
	nameField, typeField, correctValue, ranges, maxLengths := extractNameAndType(apiConfig)
	// create an array for the first test with correct value
	endpoint := apiConfig.Endpoint
	correctEndpoint := apiConfig.Endpoint
	expLength := *apiConfig.ExpectationLength
	// Check if Authentication is nil
	auth := false
	lim := false
	if apiConfig.Authentication != nil {
		auth = true
	}
	if apiConfig.Limiter != nil {
		lim = true
	}
	var minValue *int
	var MaxValue *int
	var invalidUUIDStr string
	var randomString string
	var err error
	for index, _ := range nameField {
		if typeField[index] == "int" {
			correctEndpoint = correctEndpoint + "/" + correctValue[index].(string)
			if minValue, MaxValue, err = sliptRange(*ranges[index]); err != nil {
				fmt.Println("Error: ", err)
			}
		}
		if typeField[index] == "string" {
			correctEndpoint = correctEndpoint + "/" + correctValue[index].(string)
			randomString = generateRandomString((*maxLengths[index]) + 1)
		}
		if typeField[index] == "uuid" {
			correctEndpoint = correctEndpoint + "/" + correctValue[index].(string)
			invalidUUIDStr = "not-a-valid-uuid"
		}
	}
	if lim {
		// test rate limiter with get
		generateRateLimiterGet(auth, nameFile+"_rate_lim_get_test.go", "TestRateLimiterGet"+nameFile, "GET", correctEndpoint, apiConfig.Limiter.MaxRequests, apiConfig.Limiter.Seconds)
	}
	// create first test with correct payload
	if auth {
		// se questo enpoint necessita di autenticazione anche se i parametri sono corretti dovrebbe dare come stato 401 non Autrizzato
		GenerateTestEndpointGetValue(!auth, correctEndpoint, "401", nameFile+"_gen_get_1_test.go", "TestGetCorrectValue"+nameFile, expLength)
		// nel frattempo genero il codice che necessita per ottenere il toke
		GenerateTestTokenCode()
	} else {
		GenerateTestEndpointGetValue(auth, correctEndpoint, "200", nameFile+"_gen_get_1_test.go", "TestGetCorrectValue"+nameFile, expLength)
	}

	for index, value := range typeField {
		if value == "int" {

			genericEndpoint := replaceField(correctEndpoint, endpoint, strconv.Itoa(*minValue), index)
			GenerateTestEndpointGetValue(auth, genericEndpoint, "400", nameFile+"_gen_get_"+strconv.Itoa(index)+"2_test.go", "TestGetLowerIntValue"+strconv.Itoa(index)+nameFile, expLength)

			genericEndpoint = replaceField(correctEndpoint, endpoint, strconv.Itoa(*MaxValue), index)
			GenerateTestEndpointGetValue(auth, genericEndpoint, "400", nameFile+"_gen_get_"+strconv.Itoa(index)+"3_test.go", "TestGetUpperIntValue"+strconv.Itoa(index)+nameFile, expLength)

			genericEndpoint = replaceField(correctEndpoint, endpoint, "10.1", index)
			GenerateTestEndpointGetValue(auth, genericEndpoint, "400", nameFile+"_gen_get_"+strconv.Itoa(index)+"4_test.go", "TestGetIncorrectIntValueDouble"+strconv.Itoa(index)+nameFile, expLength)

			genericEndpoint = replaceField(correctEndpoint, endpoint, "aa", index)
			GenerateTestEndpointGetValue(auth, genericEndpoint, "400", nameFile+"_gen_get_"+strconv.Itoa(index)+"5_test.go", "TestGetIncorrectIntValueString"+strconv.Itoa(index)+nameFile, expLength)

			genericEndpoint = replaceField(correctEndpoint, endpoint, "true", index)
			GenerateTestEndpointGetValue(auth, genericEndpoint, "400", nameFile+"_gen_get_"+strconv.Itoa(index)+"6_test.go", "TestGetIncorrectIntValueBoolean"+strconv.Itoa(index)+nameFile, expLength)
		}
		if value == "string" {

			genericEndpoint := replaceField(correctEndpoint, endpoint, randomString, index) //try to send a > max len string
			GenerateTestEndpointGetValue(auth, genericEndpoint, "400", nameFile+"_gen_get_"+strconv.Itoa(index)+"2_test.go", "TestGetUpperStringValue"+strconv.Itoa(index)+nameFile, expLength)

			// genericEndpoint = replaceField(correctEndpoint, endpoint, "100", index) //try to send a > max len string
			// GenerateTestEndpointGetValue(auth, genericEndpoint, "400", "gen_get_"+strconv.Itoa(index)+"3_test.go", "TestGetIncorrectStringValueInt"+strconv.Itoa(index))

			// genericEndpoint = replaceField(correctEndpoint, endpoint, "10.1", index) //try to send a > max len string
			// GenerateTestEndpointGetValue(auth, genericEndpoint, "400", "gen_get_"+strconv.Itoa(index)+"4_test.go", "TestGetIncorrectStringValueDouble"+strconv.Itoa(index))

			// genericEndpoint = replaceField(correctEndpoint, endpoint, "false", index) //try to send a > max len string
			// GenerateTestEndpointGetValue(auth, genericEndpoint, "400", "gen_get_"+strconv.Itoa(index)+"5_test.go", "TestGetIncorrectStringValueBoolean"+strconv.Itoa(index))

		}
		if value == "uuid" {
			//invalid uuid
			genericEndpoint := replaceField(correctEndpoint, endpoint, invalidUUIDStr, index) //try to send a > max len string
			GenerateTestEndpointGetValue(auth, genericEndpoint, "400", nameFile+"_gen_get_"+strconv.Itoa(index)+"2_test.go", "TestGetInvalidUuidValue"+strconv.Itoa(index)+nameFile, expLength)
		}
	}
}

func GenerateTestEndpointGetValue(auth bool, endpoint string, statusCode string, nameFile string, nameTest string, expLength int) {
	fset := token.NewFileSet()

	// Importa i pacchetti necessari

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
	importJson := &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: `"encoding/json"`,
		},
	}
	importDecl := &ast.GenDecl{}
	if statusCode == "200" {
		importDecl = &ast.GenDecl{
			Tok:   token.IMPORT,
			Specs: []ast.Spec{importTesting, importHTTP, importTest, importJson},
		}
	} else {
		importDecl = &ast.GenDecl{
			Tok:   token.IMPORT,
			Specs: []ast.Spec{importTesting, importHTTP, importTest},
		}
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
					&ast.BasicLit{Kind: token.STRING, Value: `"GET"`},
					&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("\"%s\"", endpoint)},
					ast.NewIdent("nil"),
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

	// Create the json.NewDecoder().Decode() statement
	decodeStmt := &ast.AssignStmt{
		Lhs: []ast.Expr{&ast.Ident{Name: "err"}},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "json"},
							Sel: &ast.Ident{Name: "NewDecoder"},
						},
						Args: []ast.Expr{&ast.SelectorExpr{
							X:   &ast.Ident{Name: "resp"},
							Sel: &ast.Ident{Name: "Body"},
						}},
					},
					Sel: &ast.Ident{Name: "Decode"},
				},
				Args: []ast.Expr{&ast.UnaryExpr{
					Op: token.AND,
					X:  &ast.Ident{Name: "jsonResponse"},
				}},
			},
		},
	}

	// Create the assert.Equal() statement
	assertEqualValueStmt := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "assert"},
				Sel: &ast.Ident{Name: "Equal"},
			},
			Args: []ast.Expr{
				&ast.Ident{Name: "t"},
				&ast.CallExpr{
					Fun: &ast.Ident{Name: "len"},
					Args: []ast.Expr{
						&ast.Ident{Name: "jsonResponse"},
					},
				},
				&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(expLength)},
				&ast.BasicLit{Kind: token.STRING, Value: `"Expected at least two values in the response"`},
			},
		},
	}
	jsonResponse := &ast.AssignStmt{
		Lhs: []ast.Expr{&ast.Ident{Name: "jsonResponse"}},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{&ast.CompositeLit{
			Type: &ast.ArrayType{
				Elt: &ast.Ident{Name: "map[string]interface{}"},
			},
			Elts: nil,
		}},
	}

	// make discern and if is a 200 make if correct value return is len ok
	stmts := []ast.Stmt{}
	if auth {
		if statusCode == "200" {
			stmts = []ast.Stmt{
				newRequestStmt,
				assertNoErrorStmt,
				tokenAssignment,
				firstHeader,
				authHeaderAssignment,
				doRequestStmt,
				assertNoErrorStmt,
				deferStmt,
				assertEqualStmt,
				jsonResponse,
				decodeStmt,
				assertNoErrorStmt,
				assertEqualValueStmt,
			}
		} else {
			stmts = []ast.Stmt{
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
		}
	} else {
		if statusCode == "200" {
			stmts = []ast.Stmt{
				newRequestStmt,
				assertNoErrorStmt,
				firstHeader,
				doRequestStmt,
				assertNoErrorStmt,
				deferStmt,
				assertEqualStmt,
				jsonResponse,
				decodeStmt,
				assertNoErrorStmt,
				assertEqualValueStmt,
			}
		} else {
			stmts = []ast.Stmt{
				newRequestStmt,
				assertNoErrorStmt,
				firstHeader,
				doRequestStmt,
				assertNoErrorStmt,
				deferStmt,
				assertEqualStmt,
			}
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
