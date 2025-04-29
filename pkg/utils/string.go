package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

type StringUtils struct {
}

func NewStringUtils() *StringUtils {
	return &StringUtils{}
}

func (su *StringUtils) RemoveHTMLTags(input string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(input, "")
}

func (su *StringUtils) TruncateString(input string, maxLength int) string {
	if len(input) > maxLength {
		if maxLength > 3 {
			return input[:maxLength-3] + "..."
		}
		return input[:maxLength]
	}
	return input
}

func (su *StringUtils) Trim(input string) string {
	return strings.TrimSpace(input)
}

func (su *StringUtils) GenerateRandomString(n int) (string, error) {
	if n%2 != 0 {
		return "", fmt.Errorf("length must be even to form valid hex string")
	}

	bytes := make([]byte, n/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return hex.EncodeToString(bytes), nil
}
