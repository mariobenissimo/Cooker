package models

import (
	jwt "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Username string `json:"Username"`
	jwt.RegisteredClaims
}
