package ucase

import (
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/layer/entity"
	"assistant-go/internal/layer/repository"
	"assistant-go/internal/logging"
	"assistant-go/internal/storage/postgres"
	"assistant-go/pkg/utils"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrFileTooLarge              = errors.New("file too large")
	ErrFileReading               = errors.New("error reading file")
	ErrFileInvalidType           = errors.New("invalid file type")
	ErrFileResettingPointer      = errors.New("error resetting pointer")
	ErrFileUnableToSeek          = errors.New("error unable to seek")
	ErrFileExtensionDoesNotMatch = errors.New("file extension does not match")
	ErrFileNotSafeFilename       = errors.New("file not safe filename")
	ErrFileUnableToSave          = errors.New("unable to save file")
	ErrFileSave                  = errors.New("unable to save file")
)

type FileUseCase interface {
	Upload(in dto.UploadFile, userEntity *entity.User) (*entity.File, error)
}

type fileUseCase struct {
	ctx          context.Context
	repositories *repository.Repositories
}

func NewFileUseCase(ctx context.Context, repositories *repository.Repositories) FileUseCase {
	return &fileUseCase{
		ctx:          ctx,
		repositories: repositories,
	}
}

func (uc *fileUseCase) Upload(in dto.UploadFile, userEntity *entity.User) (*entity.File, error) {
	var allowedMimeTypes = map[string][]string{
		"image/jpeg":      {".jpeg", ".jpg"},
		"image/png":       {".png"},
		"image/gif":       {".gif"},
		"application/pdf": {".pdf"},
		"application/zip": {".zip"},
	}

	limitedReader := io.LimitReader(in.File, in.MaxSizeBytes+1)

	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, err
	}

	if int64(len(data)) > in.MaxSizeBytes {
		return nil, ErrFileTooLarge
	}

	if seeker, ok := in.File.(io.Seeker); ok {
		_, err := seeker.Seek(0, io.SeekStart)
		if err != nil {
			return nil, ErrFileResettingPointer
		}
	} else {
		return nil, ErrFileUnableToSeek
	}

	buffer := make([]byte, 512)
	_, err = in.File.Read(buffer)
	if err != nil {
		return nil, ErrFileReading
	}

	mimeType := http.DetectContentType(buffer)
	extAllowed, allowed := allowedMimeTypes[mimeType]
	if !allowed {
		return nil, ErrFileInvalidType
	}

	fileExt := strings.ToLower(filepath.Ext(in.OriginalFilename))
	var extExists bool
	for _, extName := range extAllowed {
		if extName == fileExt {
			extExists = true
		}
	}
	if !extExists {
		return nil, ErrFileExtensionDoesNotMatch
	}

	safeName := filepath.Base(in.OriginalFilename)
	if strings.Contains(safeName, "..") {
		return nil, ErrFileNotSafeFilename
	}

	stringUtils := utils.NewStringUtils()
	hashForNewName, err := stringUtils.GenerateRandomString(10)
	if err != nil {
		return nil, err
	}

	newFilename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), hashForNewName, fileExt)
	filePath := filepath.Join(in.SavePath, newFilename)

	saveDto := &dto.SaveFile{
		File:      bytes.NewReader(data),
		SavePath:  filePath,
		SizeBytes: int64(len(data)),
	}

	saveErr := uc.repositories.StorageRepository.Save(saveDto)
	if saveErr != nil {
		return nil, ErrFileSave
	}

	//out, err := os.Create(filePath)
	//if err != nil {
	//	return nil, ErrFileUnableToSave
	//}
	//defer func(out *os.File) {
	//	err := out.Close()
	//	if err != nil {
	//		logging.GetLogger(uc.ctx).Errorf("error closing file: %v", err)
	//	}
	//}(out)
	//
	//_, err = io.Copy(out, in.File)
	//if err != nil {
	//	return nil, ErrFileSave
	//}

	fileHash, err := stringUtils.GenerateRandomString(100)
	if err != nil {
		return nil, err
	}

	fileEntity := &entity.File{
		UserID:           userEntity.ID,
		OriginalFilename: in.OriginalFilename,
		Filename:         newFilename,
		Ext:              fileExt,
		Size:             len(data),
		Hash:             fileHash,
		CreatedAt:        time.Now().UTC(),
	}

	_, err = uc.repositories.FileRepository.Create(fileEntity)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}

	return fileEntity, nil
}
