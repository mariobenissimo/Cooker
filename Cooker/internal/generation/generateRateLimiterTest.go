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
func generateTestFunctionBodyPost(auth bool, nameFile, nameTest string, method string, endpoint string, payload ast.AssignStmt, maxRequest int, seconds int) {
	// Create an AST node for the function body
	fset := token.NewFileSet()

	// Importa i pacchetti necessari
	importTestify := &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: `"github.com/stretchr/testify/assert"`,
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
	importTime := &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: `"time"`,
		},
	}
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
	importDecl := &ast.GenDecl{}
	if method == "POST" {
		importDecl = &ast.GenDecl{
			Tok:   token.IMPORT,
			Specs: []ast.Spec{importJSON, importBytes, importTesting, importHTTP, importTestify, importTime},
		}
	} else if method == "GET" {
		importDecl = &ast.GenDecl{
			Tok:   token.IMPORT,
			Specs: []ast.Spec{importTesting, importTestify, importHTTP, importTestify, importTime},
		}
	}
	// Add the test logic
	startTime := &ast.AssignStmt{
		Lhs: []ast.Expr{&ast.Ident{Name: "startTime"}},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{&ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "time"},
				Sel: &ast.Ident{Name: "Now"},
			},
		}},
	}

	AssignI := &ast.AssignStmt{
		Lhs: []ast.Expr{&ast.Ident{Name: "i"}},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}},
	}

	BodyFor := &ast.ForStmt{
		Cond: &ast.BinaryExpr{
			X:  &ast.Ident{Name: "i"},
			Op: token.LSS,
			Y:  &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(maxRequest)},
		},
		Post: &ast.IncDecStmt{
			X:   &ast.Ident{Name: "i"},
			Tok: token.INC,
		},
		Body: &ast.BlockStmt{
			List: generateForLoopBody(),
		},
	}

	endTime := &ast.AssignStmt{
		Lhs: []ast.Expr{&ast.Ident{Name: "endTime"}},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{&ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "time"},
				Sel: &ast.Ident{Name: "Now"},
			},
		}},
	}

	elapsedTime := &ast.AssignStmt{
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

	ifElapsed := &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  &ast.SelectorExpr{X: &ast.Ident{Name: "elapsedTime"}, Sel: &ast.Ident{Name: "Seconds()"}},
			Op: token.LSS,
			Y:  &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(seconds)},
		},
		Body: &ast.BlockStmt{
			List: generateRateLimitExceededCheck(),
		},
	}

	sleepTime := &ast.ExprStmt{
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

	// Verifica che la quarta richiesta sia consentita dopo l'attesa
	lastRequest := &ast.AssignStmt{
		Lhs: []ast.Expr{&ast.Ident{Name: "resp"}, &ast.Ident{Name: "err"}},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{&ast.CallExpr{
			Fun: &ast.Ident{Name: "makeRequest"},
		}},
	}
	lastAssertNoError := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "assert"},
				Sel: &ast.Ident{Name: "NoError"},
			},
			Args: []ast.Expr{&ast.Ident{Name: "t"}, &ast.Ident{Name: "err"}},
		},
	}
	lastAsserEqual := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "assert"},
				Sel: &ast.Ident{Name: "Equal"},
			},
			Args: []ast.Expr{&ast.Ident{Name: "t"}, &ast.BasicLit{Kind: token.INT, Value: "http.StatusOK"}, &ast.SelectorExpr{
				X:   &ast.Ident{Name: "resp"},
				Sel: &ast.Ident{Name: "StatusCode"},
			}},
		},
	}

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
		lastRequest,
		lastAssertNoError,
		lastAsserEqual,
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

func generaMakeRequestPost(auth bool, method string, endpoint string, payload ast.AssignStmt) *ast.AssignStmt {
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

	assertNoErrorStmt := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "assert"},
				Sel: &ast.Ident{Name: "NoError"},
			},
			Args: []ast.Expr{&ast.Ident{Name: "t"}, &ast.Ident{Name: "err"}},
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
	makeRequestAssignment := &ast.AssignStmt{}
	if auth {
		makeRequestAssignment = &ast.AssignStmt{
			Lhs: []ast.Expr{
				ast.NewIdent("makeRequest"),
			},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.FuncLit{
					Type: &ast.FuncType{
						Results: &ast.FieldList{
							List: []*ast.Field{
								{Names: []*ast.Ident{{Name: "*http.Response, "}}, Type: &ast.Ident{Name: "error"}},
							},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.AssignStmt{
								Lhs: []ast.Expr{&ast.Ident{Name: "client"}},
								Tok: token.DEFINE,
								Rhs: []ast.Expr{&ast.UnaryExpr{
									Op: token.AND,
									X: &ast.CompositeLit{
										Type: &ast.Ident{Name: "http.Client"},
									},
								}},
							},
							&payload,
							marshalStmt,
							assertNoErrorStmt,
							newRequestStmt,
							assertNoErrorStmt,
							firstHeader,
							tokenAssignment,
							authHeaderAssignment,
							&ast.IfStmt{
								Cond: &ast.BinaryExpr{
									X:  &ast.Ident{Name: "err"},
									Op: token.NEQ,
									Y:  ast.NewIdent("nil"),
								},
								Body: &ast.BlockStmt{
									List: []ast.Stmt{
										&ast.ReturnStmt{
											Results: []ast.Expr{ast.NewIdent("nil"), &ast.Ident{Name: "err"}},
										},
									},
								},
							},
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
				ast.NewIdent("makeRequest"),
			},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.FuncLit{
					Type: &ast.FuncType{
						Results: &ast.FieldList{
							List: []*ast.Field{
								{Names: []*ast.Ident{{Name: "*http.Response, "}}, Type: &ast.Ident{Name: "error"}},
							},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.AssignStmt{
								Lhs: []ast.Expr{&ast.Ident{Name: "client"}},
								Tok: token.DEFINE,
								Rhs: []ast.Expr{&ast.UnaryExpr{
									Op: token.AND,
									X: &ast.CompositeLit{
										Type: &ast.Ident{Name: "http.Client"},
									},
								}},
							},
							&payload,
							marshalStmt,
							assertNoErrorStmt,
							newRequestStmt,
							assertNoErrorStmt,
							firstHeader,
							&ast.IfStmt{
								Cond: &ast.BinaryExpr{
									X:  &ast.Ident{Name: "err"},
									Op: token.NEQ,
									Y:  ast.NewIdent("nil"),
								},
								Body: &ast.BlockStmt{
									List: []ast.Stmt{
										&ast.ReturnStmt{
											Results: []ast.Expr{ast.NewIdent("nil"), &ast.Ident{Name: "err"}},
										},
									},
								},
							},
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

func generateForLoopBody() []ast.Stmt {
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
				Args: []ast.Expr{&ast.Ident{Name: "t"}, &ast.BasicLit{Kind: token.INT, Value: "http.StatusOK"}, &ast.SelectorExpr{
					X:   &ast.Ident{Name: "resp"},
					Sel: &ast.Ident{Name: "StatusCode"},
				}},
			},
		},
	}
}

func generateRateLimitExceededCheck() []ast.Stmt {
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
				Args: []ast.Expr{&ast.Ident{Name: "t"}, &ast.BasicLit{Kind: token.INT, Value: "http.StatusTooManyRequests"}, &ast.SelectorExpr{
					X:   &ast.Ident{Name: "resp"},
					Sel: &ast.Ident{Name: "StatusCode"},
				}},
			},
		},
	}
}
func generateTestFunctionBodyGet(auth bool, nameFile, nameTest string, method string, endpoint string, maxRequests int, seconds int) {
	// Create an AST node for the function body
	fset := token.NewFileSet()

	// Importa i pacchetti necessari
	importTestify := &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: `"github.com/stretchr/testify/assert"`,
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
	importTime := &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: `"time"`,
		},
	}

	importDecl := &ast.GenDecl{
		Tok:   token.IMPORT,
		Specs: []ast.Spec{importTesting, importTestify, importHTTP, importTestify, importTime},
	}
	// Add the test logic
	startTime := &ast.AssignStmt{
		Lhs: []ast.Expr{&ast.Ident{Name: "startTime"}},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{&ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "time"},
				Sel: &ast.Ident{Name: "Now"},
			},
		}},
	}

	AssignI := &ast.AssignStmt{
		Lhs: []ast.Expr{&ast.Ident{Name: "i"}},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}},
	}

	BodyFor := &ast.ForStmt{
		Cond: &ast.BinaryExpr{
			X:  &ast.Ident{Name: "i"},
			Op: token.LSS,
			Y:  &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(maxRequests)},
		},
		Post: &ast.IncDecStmt{
			X:   &ast.Ident{Name: "i"},
			Tok: token.INC,
		},
		Body: &ast.BlockStmt{
			List: generateForLoopBody(),
		},
	}

	endTime := &ast.AssignStmt{
		Lhs: []ast.Expr{&ast.Ident{Name: "endTime"}},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{&ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "time"},
				Sel: &ast.Ident{Name: "Now"},
			},
		}},
	}

	elapsedTime := &ast.AssignStmt{
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

	ifElapsed := &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  &ast.SelectorExpr{X: &ast.Ident{Name: "elapsedTime"}, Sel: &ast.Ident{Name: "Seconds()"}},
			Op: token.LSS,
			Y:  &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(seconds)},
		},
		Body: &ast.BlockStmt{
			List: generateRateLimitExceededCheck(),
		},
	}

	sleepTime := &ast.ExprStmt{
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

	// Verifica che la quarta richiesta sia consentita dopo l'attesa
	lastRequest := &ast.AssignStmt{
		Lhs: []ast.Expr{&ast.Ident{Name: "resp"}, &ast.Ident{Name: "err"}},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{&ast.CallExpr{
			Fun: &ast.Ident{Name: "makeRequest"},
		}},
	}
	lastAssertNoError := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "assert"},
				Sel: &ast.Ident{Name: "NoError"},
			},
			Args: []ast.Expr{&ast.Ident{Name: "t"}, &ast.Ident{Name: "err"}},
		},
	}
	lastAsserEqual := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "assert"},
				Sel: &ast.Ident{Name: "Equal"},
			},
			Args: []ast.Expr{&ast.Ident{Name: "t"}, &ast.BasicLit{Kind: token.INT, Value: "http.StatusOK"}, &ast.SelectorExpr{
				X:   &ast.Ident{Name: "resp"},
				Sel: &ast.Ident{Name: "StatusCode"},
			}},
		},
	}

	stmts := []ast.Stmt{
		sleepTime,
		generaMakeRequestGet(auth, "GET", endpoint),
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

func generaMakeRequestGet(auth bool, method string, endpoint string) *ast.AssignStmt {
	makeRequestAssignment := &ast.AssignStmt{}
	if auth {
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

		makeRequestAssignment = &ast.AssignStmt{
			Lhs: []ast.Expr{
				ast.NewIdent("makeRequest"),
			},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.FuncLit{
					Type: &ast.FuncType{
						Results: &ast.FieldList{
							List: []*ast.Field{
								{Names: []*ast.Ident{{Name: "*http.Response, "}}, Type: &ast.Ident{Name: "error"}},
							},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.AssignStmt{
								Lhs: []ast.Expr{&ast.Ident{Name: "client"}},
								Tok: token.DEFINE,
								Rhs: []ast.Expr{&ast.UnaryExpr{
									Op: token.AND,
									X: &ast.CompositeLit{
										Type: &ast.Ident{Name: "http.Client"},
									},
								}},
							},
							&ast.AssignStmt{
								Lhs: []ast.Expr{&ast.Ident{Name: "req"}, &ast.Ident{Name: "err"}},
								Tok: token.DEFINE,
								Rhs: []ast.Expr{&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "http"},
										Sel: &ast.Ident{Name: "NewRequest"},
									},
									Args: []ast.Expr{
										&ast.BasicLit{Kind: token.STRING, Value: `"GET"`},
										&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("\"%s\"", endpoint)},
										ast.NewIdent("nil"),
									},
								}},
							},
							tokenAssignment,
							authHeaderAssignment,
							&ast.IfStmt{
								Cond: &ast.BinaryExpr{
									X:  &ast.Ident{Name: "err"},
									Op: token.NEQ,
									Y:  ast.NewIdent("nil"),
								},
								Body: &ast.BlockStmt{
									List: []ast.Stmt{
										&ast.ReturnStmt{
											Results: []ast.Expr{ast.NewIdent("nil"), &ast.Ident{Name: "err"}},
										},
									},
								},
							},
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
				ast.NewIdent("makeRequest"),
			},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.FuncLit{
					Type: &ast.FuncType{
						Results: &ast.FieldList{
							List: []*ast.Field{
								{Names: []*ast.Ident{{Name: "*http.Response, "}}, Type: &ast.Ident{Name: "error"}},
							},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.AssignStmt{
								Lhs: []ast.Expr{&ast.Ident{Name: "client"}},
								Tok: token.DEFINE,
								Rhs: []ast.Expr{&ast.UnaryExpr{
									Op: token.AND,
									X: &ast.CompositeLit{
										Type: &ast.Ident{Name: "http.Client"},
									},
								}},
							},
							&ast.AssignStmt{
								Lhs: []ast.Expr{&ast.Ident{Name: "req"}, &ast.Ident{Name: "err"}},
								Tok: token.DEFINE,
								Rhs: []ast.Expr{&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "http"},
										Sel: &ast.Ident{Name: "NewRequest"},
									},
									Args: []ast.Expr{
										&ast.BasicLit{Kind: token.STRING, Value: `"GET"`},
										&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("\"%s\"", endpoint)},
										ast.NewIdent("nil"),
									},
								}},
							},
							&ast.IfStmt{
								Cond: &ast.BinaryExpr{
									X:  &ast.Ident{Name: "err"},
									Op: token.NEQ,
									Y:  ast.NewIdent("nil"),
								},
								Body: &ast.BlockStmt{
									List: []ast.Stmt{
										&ast.ReturnStmt{
											Results: []ast.Expr{ast.NewIdent("nil"), &ast.Ident{Name: "err"}},
										},
									},
								},
							},
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
