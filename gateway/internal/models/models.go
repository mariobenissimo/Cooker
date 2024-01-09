package models

import "github.com/golang-jwt/jwt/v5"

type FavContextKey string

type Claims struct {
	Username string `json:"Username"`
	jwt.RegisteredClaims
}
