package utils

import (
	"errors"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword converts a plain text passwd into a hashed version
func HashPassword(password string) (string, error) {
	// Cost factor of 12 provides a good balance between security and performance
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", errors.New("failed to hash password")
	}
	return string(bytes), nil
}

// CheckPasswordHash compares password against hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil

}

// ValidatePassword checks password complexities requirements
func ValidatePassword(password string) error {
	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasNumber = true
		case unicode.IsSymbol(r) || unicode.IsPunct(r):
			hasSpecial = true
		}
	}

	if len(password) < 8 {
		return errors.New("password must be atleast 8 characters")
	}
	if !hasUpper {
		return errors.New("password must contain atleast an uppercase letter")
	}
	if !hasLower {
		return errors.New("password must atleast contain a lowercase letter")
	}
	if !hasNumber {
		return errors.New("password must atleast contain a digit")
	}
	if !hasSpecial {
		return errors.New("password must atleast contain a special character")
	}
	return nil
}
