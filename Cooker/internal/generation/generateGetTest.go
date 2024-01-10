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
	// create an array of numParamters which contains nameField
	// create an array of numParamters which contains type of paramters
	// create an array for the first test with correct value
	// create array with ranges (optionally) and maxLengths (optionally)
	nameField, typeField, correctValue, ranges, maxLengths := extractNameAndType(apiConfig)

	endpoint := apiConfig.Endpoint
	correctEndpoint := apiConfig.Endpoint
	expLength := *apiConfig.ExpectationLength
	// Check if Authentication is nil or lim is nil
	auth := false
	lim := false
	if apiConfig.Authentication != nil {
		auth = true
	}
	if apiConfig.Limiter != nil {
		lim = true
	}
	// inizilize variable if needed
	var minValue *int
	var MaxValue *int
	var invalidUUIDStr string
	var randomString string
	// variable for generic error
	var err error
	// create the correct endpoint
	for index, _ := range nameField {
		if typeField[index] == "int" {
			correctEndpoint = correctEndpoint + "/" + correctValue[index].(string)
			// if type is int take minValue and MaxValue form ranges
			if minValue, MaxValue, err = sliptRange(*ranges[index]); err != nil {
				fmt.Println("Error: ", err)
			}
		}
		if typeField[index] == "string" {
			correctEndpoint = correctEndpoint + "/" + correctValue[index].(string)
			// if type is string take random string with a > maxLenghts
			randomString = generateRandomString((*maxLengths[index]) + 1)
		}
		if typeField[index] == "uuid" {
			// if type is uuid take an invalid uuid
			correctEndpoint = correctEndpoint + "/" + correctValue[index].(string)
			invalidUUIDStr = "not-a-valid-uuid"
		}
	}
	if lim {
		// test rate limiter with get
		generateRateLimiterGet(auth, nameFile+"_rate_lim_get_test.go", "TestRateLimiterGet"+nameFile, "GET", correctEndpoint, apiConfig.Limiter.MaxRequests, apiConfig.Limiter.Seconds)
	}
	// if auth is true create fir
	if auth {
		// se questo enpoint necessita di autenticazione anche se i parametri sono corretti dovrebbe dare come stato 401 non Autrizzato
		GenerateTestEndpointGetValue(!auth, correctEndpoint, HTTPSTATUS_UNAUTHORIZED, nameFile+"_gen_get_0_test.go", "TestGetCorrectValue"+nameFile, expLength)
		// generate the file to get token
		GenerateTestTokenCode()
	}
	// generate the first endpoint with the correct endpoint
	GenerateTestEndpointGetValue(auth, correctEndpoint, HTTPSTATUS_OK, nameFile+"_gen_get_1_test.go", "TestGetCorrectValue"+nameFile, expLength)

	// step create combination step - fuzzy test
	for index, value := range typeField {
		if value == "int" {
			genericEndpoint := replaceField(correctEndpoint, endpoint, strconv.Itoa(*minValue), index) // min value
			GenerateTestEndpointGetValue(auth, genericEndpoint, HTTPSTATUS_BADREQUEST, nameFile+"_gen_get_"+strconv.Itoa(index)+"2_test.go", "TestGetLowerIntValue"+strconv.Itoa(index)+nameFile, expLength)

			genericEndpoint = replaceField(correctEndpoint, endpoint, strconv.Itoa(*MaxValue), index) // max value
			GenerateTestEndpointGetValue(auth, genericEndpoint, HTTPSTATUS_BADREQUEST, nameFile+"_gen_get_"+strconv.Itoa(index)+"3_test.go", "TestGetUpperIntValue"+strconv.Itoa(index)+nameFile, expLength)

			genericEndpoint = replaceField(correctEndpoint, endpoint, "10.1", index) // a double value
			GenerateTestEndpointGetValue(auth, genericEndpoint, HTTPSTATUS_BADREQUEST, nameFile+"_gen_get_"+strconv.Itoa(index)+"4_test.go", "TestGetIncorrectIntValueDouble"+strconv.Itoa(index)+nameFile, expLength)

			genericEndpoint = replaceField(correctEndpoint, endpoint, "aa", index) // a string value
			GenerateTestEndpointGetValue(auth, genericEndpoint, HTTPSTATUS_BADREQUEST, nameFile+"_gen_get_"+strconv.Itoa(index)+"5_test.go", "TestGetIncorrectIntValueString"+strconv.Itoa(index)+nameFile, expLength)

			genericEndpoint = replaceField(correctEndpoint, endpoint, "true", index) // a boolean value
			GenerateTestEndpointGetValue(auth, genericEndpoint, HTTPSTATUS_BADREQUEST, nameFile+"_gen_get_"+strconv.Itoa(index)+"6_test.go", "TestGetIncorrectIntValueBoolean"+strconv.Itoa(index)+nameFile, expLength)
		}
		if value == "string" {
			genericEndpoint := replaceField(correctEndpoint, endpoint, randomString, index) //try to send a > max len string
			GenerateTestEndpointGetValue(auth, genericEndpoint, HTTPSTATUS_BADREQUEST, nameFile+"_gen_get_"+strconv.Itoa(index)+"2_test.go", "TestGetUpperStringValue"+strconv.Itoa(index)+nameFile, expLength)
		}
		if value == "uuid" {
			genericEndpoint := replaceField(correctEndpoint, endpoint, invalidUUIDStr, index) //try to send a invalid uuid
			GenerateTestEndpointGetValue(auth, genericEndpoint, HTTPSTATUS_BADREQUEST, nameFile+"_gen_get_"+strconv.Itoa(index)+"2_test.go", "TestGetInvalidUuidValue"+strconv.Itoa(index)+nameFile, expLength)
		}
	}
}

// resp, err := http.DefaultClient.Do(req)
func CreateHTTPDefaultClient() *ast.AssignStmt {
	return &ast.AssignStmt{
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
}

// create req, err := http.NewRequest(method, endpoint, payload)
func CreateNewHTTPRequest(method, endpoint string, paylaod interface{}) *ast.AssignStmt {
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
					&ast.BasicLit{Kind: token.STRING, Value: `"GET"`},
					&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("\"%s\"", endpoint)},
					ast.NewIdent("nil"),
				},
			},
		},
	}
}

// assert.NoError(t, err)
func CreateAssertError() *ast.ExprStmt {
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "assert"},
				Sel: &ast.Ident{Name: "NoError"},
			},
			Args: []ast.Expr{&ast.Ident{Name: "t"}, &ast.Ident{Name: "err"}},
		},
	}
}
func CreateImport(nameImport string) *ast.ImportSpec {
	generateImport := &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: nameImport,
		},
	}
	return generateImport
}

// token := GetTestToken()
func CreateFunctionGetTestToken() *ast.AssignStmt {
	return &ast.AssignStmt{
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
}

// req.Header.Set(string, string)
func CreateHeader(key, value string) *ast.ExprStmt {
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.SelectorExpr{X: &ast.Ident{Name: "req"}, Sel: &ast.Ident{Name: "Header"}},
				Sel: &ast.Ident{Name: "Set"},
			},
			Args: []ast.Expr{
				&ast.BasicLit{Kind: token.STRING, Value: key},
				&ast.BasicLit{Kind: token.STRING, Value: value},
			},
		},
	}
}

// defer resp.Body.Close()
func CreateDeferBody() *ast.DeferStmt {
	return &ast.DeferStmt{
		Call: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.SelectorExpr{X: &ast.Ident{Name: "resp"}, Sel: &ast.Ident{Name: "Body"}},
				Sel: &ast.Ident{Name: "Close"},
			},
		},
	}
}

// req.Header.Set("Authorization", "Bearer "+token)
func CreateHeaderToken() *ast.ExprStmt {
	return &ast.ExprStmt{
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
}

// assert.Equal(t, statusCode, resp.StatusCode)
func CreateAssertEqualStatus(statusCode string) *ast.ExprStmt {
	return &ast.ExprStmt{
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
}

// err = json.NewDecoder(resp.Body).Decode(&jsonResponse)
func CreateJsonDecoder() *ast.AssignStmt {
	return &ast.AssignStmt{
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
}

// assert.Equal(t, len(jsonResponse), 1, "Expected at least two values in the response")
func CreateAssertEqualLen(expLength int, message string) *ast.ExprStmt {
	return &ast.ExprStmt{
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
				&ast.BasicLit{Kind: token.STRING, Value: message},
			},
		},
	}
}

// jsonResponse := []map[string]interface{}{}
func CreateJsonRespondeInterface() *ast.AssignStmt {
	return &ast.AssignStmt{
		Lhs: []ast.Expr{&ast.Ident{Name: "jsonResponse"}},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{&ast.CompositeLit{
			Type: &ast.ArrayType{
				Elt: &ast.Ident{Name: "map[string]interface{}"},
			},
			Elts: nil,
		}},
	}
}

// Create sing of the test and put inside body
func CreateTest(nameTest string, stmts []ast.Stmt) *ast.FuncDecl {
	return &ast.FuncDecl{
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
}
func CreateFile(decls []ast.Decl, packageName string) *ast.File {
	return &ast.File{
		Name:  &ast.Ident{Name: packageName},
		Decls: decls,
	}
}
func GenerateTestEndpointGetValue(auth bool, endpoint string, statusCode string, nameFile string, nameTest string, expLength int) {
	fset := token.NewFileSet()

	// Importa i pacchetti necessari

	importHTTP := CreateImport(`"net/http"`)
	importTesting := CreateImport(`"testing"`)
	importTest := CreateImport(`"github.com/stretchr/testify/assert"`)
	importJson := CreateImport(`"encoding/json"`)
	importDecl := &ast.GenDecl{}

	if statusCode == HTTPSTATUS_OK {
		importDecl = &ast.GenDecl{
			Tok:   token.IMPORT,
			Specs: []ast.Spec{importTesting, importHTTP, importTest, importJson},
		}
	} else {
		// in other case no need importJSON
		importDecl = &ast.GenDecl{
			Tok:   token.IMPORT,
			Specs: []ast.Spec{importTesting, importHTTP, importTest},
		}
	}

	// Create the http.NewRequest statement
	newRequestStmt := CreateNewHTTPRequest(`"GET"`, endpoint, nil)

	// Create the http.DefaultClient.Do statement
	doRequestStmt := CreateHTTPDefaultClient()

	// Create assert.NoError(t, err)
	assertNoErrorStmt := CreateAssertError()

	// Create token := GetTestToken()
	getTokenFunction := CreateFunctionGetTestToken()

	// Create req.Header.Set("Content-Type", "application/json")
	headerContentType := CreateHeader(`"Content-Type"`, `"application/json"`)

	// Create defer resp.Body.Close()
	deferStmt := CreateDeferBody()

	// Create req.Header.Set("Authorization", "Bearer "+token)
	authHeaderAssignment := CreateHeaderToken()

	// Create the assert.Equal statement
	assertEqualStmt := CreateAssertEqualStatus(statusCode)

	// Create the json.NewDecoder().Decode() statement
	decodeStmt := CreateJsonDecoder()

	// Create the assert.Equal() statement
	assertEqualValueStmt := CreateAssertEqualLen(expLength, `"Expected TODO in the response"`)

	// Create 	jsonResponse := []map[string]interface{}{}
	jsonResponse := CreateJsonRespondeInterface()

	// make discern and if is a 200 make if correct value return is len ok
	var stmts = []ast.Stmt{}
	if auth {
		if statusCode == HTTPSTATUS_OK {
			stmts = []ast.Stmt{
				newRequestStmt,
				assertNoErrorStmt,
				getTokenFunction,
				headerContentType,
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
				getTokenFunction,
				headerContentType,
				authHeaderAssignment,
				doRequestStmt,
				assertNoErrorStmt,
				deferStmt,
				assertEqualStmt,
			}
		}
	} else {
		if statusCode == HTTPSTATUS_OK {
			stmts = []ast.Stmt{
				newRequestStmt,
				assertNoErrorStmt,
				headerContentType,
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
				headerContentType,
				doRequestStmt,
				assertNoErrorStmt,
				deferStmt,
				assertEqualStmt,
			}
		}
	}

	// Create a string builder to hold the generated source code
	var buf strings.Builder

	// Create sign of the function and the body
	funcDecl := CreateTest(nameTest, stmts)

	decls := []ast.Decl{importDecl, funcDecl}

	// Create a new file
	file := CreateFile(decls, "test")

	// Print the generated source code to the buffer
	err := printer.Fprint(&buf, fset, file)
	if err != nil {
		fmt.Println("Error printing code:", err)
		return
	}

	// Format the source code in the buffer
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

	// Create os file and put code inside
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
