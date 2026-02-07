package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// Characters used for short ID generation (URL-safe, no ambiguous chars)
const shortIDChars = "abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// GenerateShortID generates a random short ID of the specified length
func GenerateShortID(length int) (string, error) {
	result := make([]byte, length)
	charLen := big.NewInt(int64(len(shortIDChars)))

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, charLen)
		if err != nil {
			return "", err
		}
		result[i] = shortIDChars[num.Int64()]
	}

	return string(result), nil
}

// HashFile computes SHA256 hash of data and returns hex string
func HashFile(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword compares a password with its hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// SanitizeFilename removes path traversal sequences and dangerous characters from filenames
func SanitizeFilename(filename string) string {
	// Get the base name only (removes any path components)
	filename = filepath.Base(filename)

	// Remove any null bytes
	filename = strings.ReplaceAll(filename, "\x00", "")

	// Remove leading dots (hidden files) and trailing spaces
	filename = strings.TrimLeft(filename, ".")
	filename = strings.TrimSpace(filename)

	// If filename is empty after sanitization, provide a default
	if filename == "" {
		filename = "unnamed_file"
	}

	return filename
}
