package ucase

import (
	service "assistant-go/internal/layer/service/file"
	"fmt"
	"testing"
)

func TestGetMiddlePathByFileId(t *testing.T) {
	tests := []struct {
		fileId   int
		expected string
	}{
		{fileId: 0, expected: "1/1/"},
		{fileId: 1, expected: "1/1/"},
		{fileId: 10, expected: "1/1/"},
		{fileId: 999, expected: "1/1/"},
		{fileId: 1000, expected: "1/2/"},
		{fileId: 1001, expected: "1/2/"},
		{fileId: 1999, expected: "1/2/"},
		{fileId: 2000, expected: "1/3/"},
		{fileId: 2001, expected: "1/3/"},
		{fileId: 19999, expected: "1/20/"},
		{fileId: 20000, expected: "1/21/"},
		{fileId: 20001, expected: "1/21/"},
		{fileId: 200000, expected: "1/201/"},
		{fileId: 200001, expected: "1/201/"},
		{fileId: 999999, expected: "1/1000/"},
		{fileId: 1000000, expected: "2/1/"},
		{fileId: 1000001, expected: "2/1/"},
	}

	fileService := service.NewFile().FileService()
	for _, tt := range tests {
		t.Run(fmt.Sprintf("fileId=%d", tt.fileId), func(t *testing.T) {
			result := fileService.GetMiddlePathByFileId(tt.fileId)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}
