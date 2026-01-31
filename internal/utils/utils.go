package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"math/big"

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

// HashFileReader computes SHA256 hash from a reader and returns hex string
func HashFileReader(r io.Reader) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
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
