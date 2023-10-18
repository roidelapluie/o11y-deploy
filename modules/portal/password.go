package portal

import (
	"crypto/rand"
	"encoding/hex"
	"io/ioutil"
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
)

const (
	passwordLength = 10
	passwordFile   = "admin-password"
)

func generatePassword(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}

func hashAndSavePassword(password, filePath string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), ioutil.WriteFile(filePath, hashedPassword, 0600)
}

func getBcryptFromFile(filePath string) (string, error) {
	hash, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func getOrGetPassword(dataDir string) (string, string, error) {
	filePath := filepath.Join(dataDir, passwordFile)

	// Check if the file already exists.
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// File does not exist, generate a new password.
		password, err := generatePassword(passwordLength)
		if err != nil {
			return "", "", err
		}

		// Ensure the dataDir exists.
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return "", "", err
		}

		// Hash and save the password.
		hash, err := hashAndSavePassword(password, filePath)
		if err != nil {
			return "", "", err
		}
		return password, hash, nil
	} else {
		// File exists, read the hash.
		hash, err := getBcryptFromFile(filePath)
		if err != nil {
			return "", "", err
		}
		return "", hash, nil
	}
}
