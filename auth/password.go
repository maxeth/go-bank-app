package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hpw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash pw %w", err)
	}

	return string(hpw), nil
}

func CheckPassword(hPw, pw string) error {
	return bcrypt.CompareHashAndPassword([]byte(hPw), []byte(pw))
}
