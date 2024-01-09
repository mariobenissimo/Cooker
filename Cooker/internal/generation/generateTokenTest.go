package generation

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"
	"go/token"
	"os"
	"strings"
)

func GenerateTestTokenCode() {
	// Create a new file set
	fset := token.NewFileSet()

	// Import necessary packages
	importDecls := &ast.GenDecl{
		Tok: token.IMPORT,
		Specs: []ast.Spec{
			&ast.ImportSpec{
				Path: &ast.BasicLit{
					Kind:  token.STRING,
					Value: `"time"`,
				},
			},
			&ast.ImportSpec{
				Path: &ast.BasicLit{
					Kind:  token.STRING,
					Value: `"github.com/golang-jwt/jwt/v5"`,
				},
			},
			&ast.ImportSpec{
				Path: &ast.BasicLit{
					Kind:  token.STRING,
					Value: `"github.com/mariobenissimo/Cooker/internal/models"`,
				},
			},
		},
	}

	// Define the GetTestToken function body
	funcDecl := &ast.FuncDecl{
		Name: ast.NewIdent("GetTestToken"),
		Type: &ast.FuncType{
			Results: &ast.FieldList{List: []*ast.Field{
				{
					Names: []*ast.Ident{{Name: "stringResult"}},
					Type:  &ast.Ident{Name: "string"},
				},
			}},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.AssignStmt{
					Lhs: []ast.Expr{
						&ast.Ident{Name: "jwtKey"},
					},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.ArrayType{
								Elt: &ast.Ident{Name: "byte"},
							},
							Args: []ast.Expr{
								&ast.BasicLit{Kind: token.STRING, Value: "`secret`"},
							},
						},
					},
				},
				&ast.AssignStmt{
					Lhs: []ast.Expr{
						&ast.Ident{Name: "expirationTime"},
					},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X: &ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "time"},
										Sel: &ast.Ident{Name: "Now"},
									},
								},
								Sel: &ast.Ident{Name: "Add"},
							},
							Args: []ast.Expr{
								&ast.BinaryExpr{
									X:  &ast.BasicLit{Kind: token.INT, Value: "5"},
									Op: token.MUL,
									Y: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "time"},
										Sel: &ast.Ident{Name: "Minute"},
									},
								},
							},
						},
					},
				},
				&ast.AssignStmt{
					Lhs: []ast.Expr{
						&ast.Ident{Name: "claims"},
					},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.UnaryExpr{
							Op: token.AND,
							X: &ast.CompositeLit{
								Type: &ast.SelectorExpr{
									X:   &ast.Ident{Name: "models"},
									Sel: &ast.Ident{Name: "Claims"},
								},
								Elts: []ast.Expr{
									&ast.KeyValueExpr{
										Key:   &ast.Ident{Name: "Username"},
										Value: &ast.BasicLit{Kind: token.STRING, Value: "\"mario\""},
									},
									&ast.KeyValueExpr{
										Key: &ast.Ident{Name: "RegisteredClaims"},
										Value: &ast.CompositeLit{
											Type: &ast.SelectorExpr{X: &ast.Ident{Name: "jwt"}, Sel: &ast.Ident{Name: "RegisteredClaims"}},
											Elts: []ast.Expr{&ast.KeyValueExpr{
												Key:   &ast.Ident{Name: "ExpiresAt"},
												Value: &ast.CallExpr{Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "jwt"}, Sel: &ast.Ident{Name: "NewNumericDate"}}, Args: []ast.Expr{&ast.Ident{Name: "expirationTime"}}},
											},
											},
										},
									},
								},
							},
						},
					},
				},
				&ast.AssignStmt{
					Lhs: []ast.Expr{
						&ast.Ident{Name: "token"},
					},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   &ast.Ident{Name: "jwt"},
								Sel: &ast.Ident{Name: "NewWithClaims"},
							},
							Args: []ast.Expr{
								&ast.SelectorExpr{
									X:   &ast.Ident{Name: "jwt"},
									Sel: &ast.Ident{Name: "SigningMethodHS256"},
								},
								&ast.Ident{Name: "claims"},
							},
						},
					},
				},
				&ast.AssignStmt{
					Lhs: []ast.Expr{
						&ast.Ident{Name: "tokenString"},
						&ast.Ident{Name: "_"},
					},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   &ast.Ident{Name: "token"},
								Sel: &ast.Ident{Name: "SignedString"},
							},
							Args: []ast.Expr{
								&ast.Ident{Name: "jwtKey"},
							},
						},
					},
				},
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.Ident{Name: "tokenString"},
					},
				},
			},
		},
	}

	// Add declarations to the list of declarations
	decls := []ast.Decl{importDecls, funcDecl}

	// Create a string builder to hold the generated source code
	var buf strings.Builder

	// Create a new file
	file := &ast.File{
		Name:  &ast.Ident{Name: "main"},
		Decls: decls,
	}

	// Print the generated source code to the buffer
	err := printer.Fprint(&buf, fset, file)
	if err != nil {
		fmt.Println("error printing code:", err)
		return
	}
	// Format the source code in the buffer
	formattedCode, err := format.Source([]byte(buf.String()))
	if err != nil {
		fmt.Println("error formatting code:", err)
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
	filePath := folderPath + "/gen_token.go"

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
	fmt.Println("Generated test code written to gen_token")
}
