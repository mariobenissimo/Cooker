package test

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/mariobenissimo/Cooker/internal/models"
	"time"
)

func GetTestToken() (stringResult string) {
	jwtKey := []byte("secret")
	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &models.Claims{Username: "mario", RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(expirationTime)}}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(jwtKey)
	return tokenString
}
