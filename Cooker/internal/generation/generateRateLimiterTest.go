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
)

func generateRateLimiterGet(auth bool, nameFile string, nameTest string, method string, endpoint string, maxRequest int, seconds int) {
	generateTestFunctionBodyGet(auth, nameFile, nameTest, method, endpoint, maxRequest, seconds)
}
func generateRateLimiterPost(auth bool, nameFile, nameTest string, method string, endpoint string, payload ast.AssignStmt, maxRequest int, seconds int) {
	generateTestFunctionBodyPost(auth, nameFile, nameTest, method, endpoint, payload, maxRequest, seconds)
}
func CreateStartTimeStm() *ast.AssignStmt {
	return &ast.AssignStmt{
		Lhs: []ast.Expr{&ast.Ident{Name: "startTime"}},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{&ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "time"},
				Sel: &ast.Ident{Name: "Now"},
			},
		}},
	}
}
func CreateBasicStm(key, value string) *ast.AssignStmt {
	return &ast.AssignStmt{
		Lhs: []ast.Expr{&ast.Ident{Name: key}},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: value}},
	}
}
func CreateFor(key string, statusCode string, maxRequest int) *ast.ForStmt {
	return &ast.ForStmt{
		Cond: &ast.BinaryExpr{
			X:  &ast.Ident{Name: key},
			Op: token.LSS,
			Y:  &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(maxRequest)},
		},
		Post: &ast.IncDecStmt{
			X:   &ast.Ident{Name: key},
			Tok: token.INC,
		},
		Body: &ast.BlockStmt{
			List: generateForLoopBody(statusCode),
		},
	}
}

// endTime := time.Now()
func CreateEndTime() *ast.AssignStmt {
	return &ast.AssignStmt{
		Lhs: []ast.Expr{&ast.Ident{Name: "endTime"}},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{&ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "time"},
				Sel: &ast.Ident{Name: "Now"},
			},
		}},
	}
}

// elapsedTime := endTime.Sub(startTime)
func CreateElapsedTime() *ast.AssignStmt {
	return &ast.AssignStmt{
		Lhs: []ast.Expr{&ast.Ident{Name: "elapsedTime"}},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{&ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "endTime"},
				Sel: &ast.Ident{Name: "Sub"},
			},
			Args: []ast.Expr{&ast.Ident{Name: "startTime"}},
		}},
	}
}

func CreateIfElapsed(seconds int) *ast.IfStmt {
	return &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  &ast.SelectorExpr{X: &ast.Ident{Name: "elapsedTime"}, Sel: &ast.Ident{Name: "Seconds()"}},
			Op: token.LSS,
			Y:  &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(seconds)},
		},
		Body: &ast.BlockStmt{
			List: CreateRequest(HTTPSTATUS_TOOMANYREQUEST),
		},
	}
}

// time.Sleep(seconds * time.Second)
func CreateSleep(seconds int) *ast.ExprStmt {
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "time"},
				Sel: &ast.Ident{Name: "Sleep"},
			},
			Args: []ast.Expr{&ast.BinaryExpr{
				X:  &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(seconds)},
				Op: token.MUL,
				Y: &ast.SelectorExpr{
					X:   ast.NewIdent("time"),
					Sel: ast.NewIdent("Second"),
				},
			}},
		},
	}
}

// resp, err := makeRequest()
//
//	assert.NoError(t, err)
//	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)

func CreateRequest(statusCode string) []ast.Stmt {
	return []ast.Stmt{
		&ast.AssignStmt{
			Lhs: []ast.Expr{&ast.Ident{Name: "resp"}, &ast.Ident{Name: "err"}},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.CallExpr{
				Fun: &ast.Ident{Name: "makeRequest"},
			}},
		},
		&ast.ExprStmt{
			X: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "assert"},
					Sel: &ast.Ident{Name: "NoError"},
				},
				Args: []ast.Expr{&ast.Ident{Name: "t"}, &ast.Ident{Name: "err"}},
			},
		},
		&ast.ExprStmt{
			X: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "assert"},
					Sel: &ast.Ident{Name: "Equal"},
				},
				Args: []ast.Expr{&ast.Ident{Name: "t"}, &ast.BasicLit{Kind: token.INT, Value: statusCode}, &ast.SelectorExpr{
					X:   &ast.Ident{Name: "resp"},
					Sel: &ast.Ident{Name: "StatusCode"},
				}},
			},
		},
	}
}

// client := &http.Client{}
func CreateClient() *ast.AssignStmt {
	return &ast.AssignStmt{
		Lhs: []ast.Expr{&ast.Ident{Name: "client"}},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{&ast.UnaryExpr{
			Op: token.AND,
			X: &ast.CompositeLit{
				Type: &ast.Ident{Name: "http.Client"},
			},
		}},
	}
}
func generaMakeRequestPost(auth bool, method string, endpoint string, payload ast.AssignStmt) *ast.AssignStmt {

	marshalStmt := CreateJsonRequest()

	// Create the http.NewRequest statement
	newRequestStmt := CreateNewHTTPRequestPayload(`"POST"`, endpoint)

	assertNoErrorStmt := CreateAssertError()

	firstHeader := CreateHeader(`"Content-Type"`, `"application/json"`)

	tokenAssignment := CreateFunctionGetTestToken()

	authHeaderAssignment := CreateHeaderToken()

	makeRequestAssignment := &ast.AssignStmt{}

	client := CreateClient()

	//start assembling makeRequest funciton
	if auth {
		makeRequestAssignment = &ast.AssignStmt{
			Lhs: []ast.Expr{
				ast.NewIdent("makeRequest"), // name function
			},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.FuncLit{
					Type: &ast.FuncType{
						Results: &ast.FieldList{
							List: []*ast.Field{
								{Names: []*ast.Ident{{Name: "*http.Response, "}}, Type: &ast.Ident{Name: "error"}}, // return type
							},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							client,
							&payload,
							marshalStmt,
							assertNoErrorStmt,
							newRequestStmt,
							assertNoErrorStmt,
							firstHeader,
							tokenAssignment,
							authHeaderAssignment,
							assertNoErrorStmt,
							&ast.ReturnStmt{
								Results: []ast.Expr{&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "client"},
										Sel: &ast.Ident{Name: "Do"},
									},
									Args: []ast.Expr{ast.NewIdent("req")},
								}},
							},
						},
					},
				},
			},
		}
	} else {
		makeRequestAssignment = &ast.AssignStmt{
			Lhs: []ast.Expr{
				ast.NewIdent("makeRequest"), // name function
			},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.FuncLit{
					Type: &ast.FuncType{
						Results: &ast.FieldList{
							List: []*ast.Field{
								{Names: []*ast.Ident{{Name: "*http.Response, "}}, Type: &ast.Ident{Name: "error"}}, // return type
							},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							client,
							&payload,
							marshalStmt,
							assertNoErrorStmt,
							newRequestStmt,
							assertNoErrorStmt,
							firstHeader,
							assertNoErrorStmt,
							&ast.ReturnStmt{
								Results: []ast.Expr{&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "client"},
										Sel: &ast.Ident{Name: "Do"},
									},
									Args: []ast.Expr{ast.NewIdent("req")},
								}},
							},
						},
					},
				},
			},
		}
	}
	return makeRequestAssignment
}
func generateTestFunctionBodyPost(auth bool, nameFile, nameTest string, method string, endpoint string, payload ast.AssignStmt, maxRequest int, seconds int) {

	fset := token.NewFileSet()

	// Package needed
	importTestify := CreateImport(`"github.com/stretchr/testify/assert"`)

	importHTTP := CreateImport(`"net/http"`)

	importTesting := CreateImport(`"testing"`)

	importTime := CreateImport(`"time"`)

	importBytes := CreateImport(`"bytes"`)

	importJSON := CreateImport(`"encoding/json"`)

	importDecl := &ast.GenDecl{}

	importDecl = &ast.GenDecl{
		Tok:   token.IMPORT,
		Specs: []ast.Spec{importJSON, importBytes, importTesting, importHTTP, importTestify, importTime},
	}

	// Create startTime := time.Now()
	startTime := CreateStartTimeStm()

	// Create i := 0
	AssignI := CreateBasicStm("i", "0")

	// Create for and his body
	BodyFor := CreateFor("i", HTTPSTATUS_CREATED, maxRequest)

	// Create endTime := time.Now()
	endTime := CreateEndTime()

	// elapsedTime := endTime.Sub(startTime)
	elapsedTime := CreateElapsedTime()

	// if elapsedTime.Seconds() < seconds
	ifElapsed := CreateIfElapsed(seconds)

	// time.Sleep(seconds * time.Second)
	sleepTime := CreateSleep(seconds)

	// Create last request
	lastRequest := CreateRequest(HTTPSTATUS_CREATED)

	lastMakeRequest := lastRequest[0]

	assertNoError := lastRequest[1]

	assertEqual := lastRequest[2]

	stmts := []ast.Stmt{
		sleepTime,
		generaMakeRequestPost(auth, method, endpoint, payload),
		startTime,
		AssignI,
		BodyFor,
		endTime,
		elapsedTime,
		ifElapsed,
		sleepTime,
		lastMakeRequest,
		assertNoError,
		assertEqual,
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

// FINISH POST RATE LIMITER

// START GET RATE LIMITER
func generateForLoopBody(statusCode string) []ast.Stmt {
	return []ast.Stmt{
		&ast.AssignStmt{
			Lhs: []ast.Expr{&ast.Ident{Name: "resp"}, &ast.Ident{Name: "err"}},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.CallExpr{
				Fun: &ast.Ident{Name: "makeRequest"},
			}},
		},
		&ast.ExprStmt{
			X: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "assert"},
					Sel: &ast.Ident{Name: "NoError"},
				},
				Args: []ast.Expr{&ast.Ident{Name: "t"}, &ast.Ident{Name: "err"}},
			},
		},
		&ast.ExprStmt{
			X: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "assert"},
					Sel: &ast.Ident{Name: "Equal"},
				},
				Args: []ast.Expr{&ast.Ident{Name: "t"}, &ast.BasicLit{Kind: token.INT, Value: statusCode}, &ast.SelectorExpr{
					X:   &ast.Ident{Name: "resp"},
					Sel: &ast.Ident{Name: "StatusCode"},
				}},
			},
		},
	}
}
func CreateForGet(key string, statusCode string, maxRequests int) *ast.ForStmt {

	return &ast.ForStmt{
		Cond: &ast.BinaryExpr{
			X:  &ast.Ident{Name: key},
			Op: token.LSS,
			Y:  &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(maxRequests)},
		},
		Post: &ast.IncDecStmt{
			X:   &ast.Ident{Name: key},
			Tok: token.INC,
		},
		Body: &ast.BlockStmt{
			List: generateForLoopBody(statusCode),
		},
	}
}
func generateTestFunctionBodyGet(auth bool, nameFile, nameTest string, method string, endpoint string, maxRequests int, seconds int) {
	// Create an AST node for the function body
	fset := token.NewFileSet()

	// Importa i pacchetti necessari
	importTestify := CreateImport(`"github.com/stretchr/testify/assert"`)

	importHTTP := CreateImport(`"net/http"`)

	importTesting := CreateImport(`"testing"`)

	importTime := CreateImport(`"time"`)

	importDecl := &ast.GenDecl{
		Tok:   token.IMPORT,
		Specs: []ast.Spec{importTesting, importTestify, importHTTP, importTestify, importTime},
	}
	// Add the test logic
	startTime := CreateStartTimeStm()

	AssignI := CreateBasicStm("i", "0")

	BodyFor := CreateForGet("i", HTTPSTATUS_OK, maxRequests)

	endTime := CreateEndTime()

	elapsedTime := CreateElapsedTime()

	ifElapsed := CreateIfElapsed(seconds)

	sleepTime := CreateSleep(seconds)

	// Create last request
	lastCompleteRequest := CreateRequest(HTTPSTATUS_OK)

	lastRequest := lastCompleteRequest[0]

	lastAssertNoError := lastCompleteRequest[1]

	lastAsserEqual := lastCompleteRequest[2]

	stmts := []ast.Stmt{
		sleepTime,
		generaMakeRequestGet(auth, `"GET"`, endpoint),
		startTime,
		AssignI,
		BodyFor,
		endTime,
		elapsedTime,
		ifElapsed,
		sleepTime,
		lastRequest,
		lastAssertNoError,
		lastAsserEqual,
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

func generaMakeRequestGet(auth bool, method string, endpoint string) *ast.AssignStmt {

	// Create token := GetTestToken()
	tokenAssignment := CreateFunctionGetTestToken()
	// Create req.Header.Set("Authorization", "Bearer "+token)
	authHeaderAssignment := CreateHeaderToken()
	client := CreateClient()
	// Create the http.NewRequest statement
	newRequestStmt := CreateNewHTTPRequest("GET", endpoint, nil)
	// Create assert.NoError(t, err)
	assertNoErrorStmt := CreateAssertError()

	makeRequestAssignment := &ast.AssignStmt{}
	if auth {
		makeRequestAssignment = &ast.AssignStmt{
			Lhs: []ast.Expr{
				ast.NewIdent("makeRequest"), // name of the function
			},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.FuncLit{
					Type: &ast.FuncType{
						Results: &ast.FieldList{
							List: []*ast.Field{
								{Names: []*ast.Ident{{Name: "*http.Response, "}}, Type: &ast.Ident{Name: "error"}}, // return type
							},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							client,
							newRequestStmt,
							tokenAssignment,
							authHeaderAssignment,
							assertNoErrorStmt,
							&ast.ReturnStmt{
								Results: []ast.Expr{&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "client"},
										Sel: &ast.Ident{Name: "Do"},
									},
									Args: []ast.Expr{ast.NewIdent("req")},
								}},
							},
						},
					},
				},
			},
		}
	} else { // case of no auth
		makeRequestAssignment = &ast.AssignStmt{
			Lhs: []ast.Expr{
				ast.NewIdent("makeRequest"), //name funtion
			},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.FuncLit{
					Type: &ast.FuncType{
						Results: &ast.FieldList{
							List: []*ast.Field{
								{Names: []*ast.Ident{{Name: "*http.Response, "}}, Type: &ast.Ident{Name: "error"}}, //return type
							},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							client,
							newRequestStmt,
							assertNoErrorStmt,
							&ast.ReturnStmt{
								Results: []ast.Expr{&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "client"},
										Sel: &ast.Ident{Name: "Do"},
									},
									Args: []ast.Expr{ast.NewIdent("req")},
								}},
							},
						},
					},
				},
			},
		}
	}
	return makeRequestAssignment
}
