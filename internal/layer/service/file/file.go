package service

import (
	"assistant-go/pkg/utils"
	"fmt"
	"time"
)

type FileService interface {
	GetMiddlePathByFileId(fileId int) string
	GenerateNewFileName(fileExt string) (string, error)
	GenerateFileHash() (string, error)
}

type fileService struct{}

func (s *fileService) GetMiddlePathByFileId(fileId int) string {
	directoryLevel1 := fileId / 1_000_000
	directoryLevel2 := (fileId % 1_000_000) / 1_000
	return fmt.Sprintf("%d/%d/", directoryLevel1+1, directoryLevel2+1)
}

func (s *fileService) GenerateNewFileName(fileExt string) (string, error) {
	stringUtils := utils.NewStringUtils()
	hashForNewName, err := stringUtils.GenerateRandomString(10)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d_%s.%s", time.Now().UnixNano(), hashForNewName, fileExt), nil
}

func (s *fileService) GenerateFileHash() (string, error) {
	stringUtils := utils.NewStringUtils()
	fileHash, err := stringUtils.GenerateRandomString(80)
	if err != nil {
		return "", err
	}
	return fileHash, nil
}
