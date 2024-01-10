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

// jwtKey := []byte(`secret`)
func CreateJwtKey(secret string) *ast.AssignStmt {
	return &ast.AssignStmt{
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
					&ast.BasicLit{Kind: token.STRING, Value: secret},
				},
			},
		},
	}
}

// expirationTime := time.Now().Add(5 * time.Minute)
func CreateExpirationTime(minutes string) *ast.AssignStmt {
	return &ast.AssignStmt{
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
						X:  &ast.BasicLit{Kind: token.INT, Value: minutes},
						Op: token.MUL,
						Y: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "time"},
							Sel: &ast.Ident{Name: "Minute"},
						},
					},
				},
			},
		},
	}
}
func CreateClaims(key, value string) *ast.AssignStmt {
	return &ast.AssignStmt{
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
							Key:   &ast.Ident{Name: key},
							Value: &ast.BasicLit{Kind: token.STRING, Value: value},
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
	}
}

// token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
func CreateToken() *ast.AssignStmt {
	return &ast.AssignStmt{
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
	}
}

// tokenString, _ := token.SignedString(jwtKey)
func CreateSignedToken() *ast.AssignStmt {
	return &ast.AssignStmt{
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
	}
}
func GenerateTestTokenCode() {
	// Create a new file set
	fset := token.NewFileSet()

	// Import necessary packages
	importTime := CreateImport(`"time"`)

	importJwt := CreateImport(`"github.com/golang-jwt/jwt/v5"`)

	importModels := CreateImport(`"github.com/mariobenissimo/Cooker/internal/models"`)

	importDecls := &ast.GenDecl{
		Tok:   token.IMPORT,
		Specs: []ast.Spec{importTime, importJwt, importModels},
	}

	// jwtKey := []byte(`secret`)
	jwtKey := CreateJwtKey(`"secret"`)

	// expirationTime := time.Now().Add(5 * time.Minute)
	expirationTime := CreateExpirationTime("5")

	// 	claims := &models.Claims{Username: "mario", RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(expirationTime)}}
	claims := CreateClaims("Username", `"mario"`)

	//	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token := CreateToken()

	// 	tokenString, _ := token.SignedString(jwtKey)
	tokenString := CreateSignedToken()

	stmts := []ast.Stmt{
		jwtKey,
		expirationTime,
		claims,
		token,
		tokenString,
		&ast.ReturnStmt{
			Results: []ast.Expr{
				&ast.Ident{Name: "tokenString"},
			},
		},
	}
	funcDecl := &ast.FuncDecl{
		Name: ast.NewIdent("GetTestToken"), // name file
		Type: &ast.FuncType{
			Results: &ast.FieldList{List: []*ast.Field{
				{
					Names: []*ast.Ident{{Name: "stringResult"}}, //return type
					Type:  &ast.Ident{Name: "string"},
				},
			}},
		},
		Body: &ast.BlockStmt{
			List: stmts,
		},
	}

	// Add declarations to the list of declarations
	decls := []ast.Decl{importDecls, funcDecl}

	// Create a string builder to hold the generated source code
	var buf strings.Builder

	// Create a new file
	file := CreateFile(decls, "test")

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
	filePath := folderPath + "/gen_token.go"

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
	fmt.Println("Generated test code written to gen_token")
}
