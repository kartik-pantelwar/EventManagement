package utilities

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// create hashfunction
func HashPassword(password string) (string, error) {
	//code
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash passsword: %w", err)
	}
	return string(hashedPassword), err
}

func CheckPassword(hashedPassword string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
