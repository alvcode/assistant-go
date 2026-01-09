package service

import (
	"assistant-go/internal/layer/dto"
	"assistant-go/pkg/utils"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"mime/multipart"
	"time"
)

type FileService interface {
	GetMiddlePathByFileId(fileId int) string
	GenerateNewFileName(fileExt string) (string, error)
	GenerateFileHash() (string, error)
	EncryptFile(file multipart.File, encryptionKey string) (multipart.File, error)
	DecryptFile(file io.Reader, encryptionKey string) (io.Reader, error)
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

func (s *fileService) EncryptFile(file multipart.File, encryptionKey string) (multipart.File, error) {
	plaintext, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(s.deriveAESKeyFromEnv(encryptionKey))
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)

	buf := make([]byte, 0, len(nonce)+len(ciphertext))
	buf = append(buf, nonce...)
	buf = append(buf, ciphertext...)

	return &dto.MemoryMultipartFile{
		Reader: bytes.NewReader(buf),
	}, nil
}

func (s *fileService) DecryptFile(file io.Reader, encryptionKey string) (io.Reader, error) {
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(s.deriveAESKeyFromEnv(encryptionKey))
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesgcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("data too short: cannot contain nonce")
	}

	nonce := data[:nonceSize]
	ciphertext := data[nonceSize:]

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(plaintext), nil
}

func (s *fileService) deriveAESKeyFromEnv(envKey string) []byte {
	hash := sha256.Sum256([]byte(envKey))
	return hash[:]
}
