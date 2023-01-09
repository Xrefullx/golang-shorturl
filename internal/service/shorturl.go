package service

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"
)

func sha256Of(input string) []byte {
	algorithm := sha256.New()
	algorithm.Write([]byte(input))
	return algorithm.Sum(nil)
}

func GenerateLink(searchlink string, hash string) string {
	url := sha256Of(searchlink + hash)
	encod := base64.StdEncoding.EncodeToString(url)
	slashReplaced := strings.ReplaceAll(encod[:8], "/", "+")
	return encod[:8]

	return slashReplaced
}
