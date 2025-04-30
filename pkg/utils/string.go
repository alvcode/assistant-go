package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
	"strings"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

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
	if n <= 0 {
		return "", fmt.Errorf("length must be positive")
	}

	b := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letterBytes))))
		if err != nil {
			return "", err
		}
		b[i] = letterBytes[num.Int64()]
	}

	return string(b), nil
}
