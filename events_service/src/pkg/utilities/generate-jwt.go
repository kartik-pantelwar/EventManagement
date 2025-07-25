package utilities

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Claims struct {
	Uid  int    `json:"uid"`
	Role string `json:"role"`
	jwt.StandardClaims
}

var jwtKey = []byte("kfladsoifdwfds")

func GenerateJWT(uid int, role string) (string, time.Time, error) {
	expirationTime := time.Now().Add(5 * time.Hour)
	claims := &Claims{
		Uid:  uid,
		Role: role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	//! I was using Signing Method ES256 instead of HS256
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", time.Now(), err
	}
	return tokenString, expirationTime, nil
}

func ValidateJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return nil, err
		}
		return nil, err
	}
	if !token.Valid {
		return nil, err
	}
	return claims, nil
}
