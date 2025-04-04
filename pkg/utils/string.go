package utils

import (
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
