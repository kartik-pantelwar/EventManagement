package utilities

import (
	"authservice/src/internal/core/session"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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

func GenerateSession(userId int) (session.Session, error) {
	tokenID := uuid.New()
	expiresAt := time.Now().Add(20 * time.Hour) //!Default was 2 * time.Hour
	issuedAt := time.Now()

	hashToken, err := bcrypt.GenerateFromPassword([]byte(tokenID.String()), bcrypt.DefaultCost)
	if err != nil {
		return session.Session{}, err
	}

	session := session.Session{
		Id:        tokenID,
		Uid:       userId,
		TokenHash: string(hashToken),
		ExpiresAt: expiresAt,
		IssuedAt:  issuedAt,
	}
	return session, nil
}
