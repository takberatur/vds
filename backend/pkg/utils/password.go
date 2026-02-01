package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sethvargo/go-password/password"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
func ComparePasswords(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
func GenerateRandomPassword() string {
	strRes, err := password.Generate(64, 10, 10, false, false)
	if err != nil {
		return GenerateRandomSecureString(64)
	}
	return strRes
}
func GenerateRandomSecureString(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return GenerateRandomToken()
	}
	return base64.URLEncoding.EncodeToString(b)
}
func GenerateRandomToken() string {
	uuidStr := uuid.New().String()
	timestamp := time.Now().UnixNano()
	uuidWithTImestamp := fmt.Sprintf("%s%x", uuidStr[:24], timestamp)
	return uuidWithTImestamp
}

func GenerateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}
	return string(bytes), nil
}
