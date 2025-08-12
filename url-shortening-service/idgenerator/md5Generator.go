package idgenerator

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"strings"
)

// Compile-time check that Md5Generator implements type IDGeneratorInterface interface {
var _ IDGeneratorInterface = (*Md5Generator)(nil)

type Md5Generator struct {
	shortCodeLength int
}

func NewMD5Generator(shortCodeLength int) *Md5Generator {
	return &Md5Generator{
		shortCodeLength: shortCodeLength,
	}
}

// md5 hash generation with base64 conversion
func (g *Md5Generator) GenerateShortCode() (string, error) {
	// Generate random bytes
	bytes := make([]byte, g.shortCodeLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	hasher := md5.New()
	hasher.Write([]byte(bytes))
	hash := hasher.Sum(nil)

	// convert to base 64 helps so same number between
	// its different number representation systems
	// also makes URL-safe
	encoded := base64.URLEncoding.EncodeToString(hash)

	// Take first 7 characters and remove any special chars
	shortCode := strings.ReplaceAll(encoded[:g.shortCodeLength], "/", "_")
	shortCode = strings.ReplaceAll(shortCode, "+", "-")
	shortCode = strings.ReplaceAll(shortCode, "=", "")

	return shortCode, nil
}
