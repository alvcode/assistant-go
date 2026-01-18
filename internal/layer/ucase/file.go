package ucase

import (
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/layer/entity"
	"assistant-go/internal/layer/repository"
	service "assistant-go/internal/layer/service/file"
	"assistant-go/internal/logging"
	"assistant-go/internal/storage/postgres"
	"bytes"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
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
	ErrFileSave                  = errors.New("unable to save file")
	ErrFileNotFound              = errors.New("file not found")
	ErrFileSystemIsFull          = errors.New("file system is full")
)

type FileUseCase interface {
	Upload(ctx context.Context, in dto.UploadFile, userEntity *entity.User) (*entity.File, error)
	GetFileByHash(ctx context.Context, in dto.GetFileByHash) (*dto.FileResponse, error)
	DeleteByID(ctx context.Context, fileID int, generalPath string) error
	CleanUnused(ctx context.Context, generalPath string) error
	GetAllowedMimeTypes() map[string][]string
	GetAllowedExtensions() []string
}

type fileUseCase struct {
	repositories *repository.Repositories
}

func NewFileUseCase(repositories *repository.Repositories) FileUseCase {
	return &fileUseCase{
		repositories: repositories,
	}
}

func (uc *fileUseCase) GetAllowedMimeTypes() map[string][]string {
	return map[string][]string{
		"image/jpeg":         {".jpeg", ".jpg"},
		"image/png":          {".png"},
		"image/gif":          {".gif"},
		"application/pdf":    {".pdf"},
		"application/zip":    {".zip", ".xlsx", ".docx"},
		"application/msword": {".doc"},
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": {".docx"},
		"application/vnd.ms-excel": {".xls"},
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": {".xlsx"},
		"application/octet-stream": {".doc", ".docx", ".xls", ".xlsx"},
	}
}

func (uc *fileUseCase) GetAllowedExtensions() []string {
	seen := make(map[string]struct{})
	var result []string
	for _, extSlice := range uc.GetAllowedMimeTypes() {
		for _, extension := range extSlice {
			if _, ok := seen[extension]; !ok {
				seen[extension] = struct{}{}
				result = append(result, extension)
			}
		}
	}
	return result
}

func (uc *fileUseCase) Upload(ctx context.Context, in dto.UploadFile, userEntity *entity.User) (*entity.File, error) {
	fileService := service.NewFile().FileService()

	limitedReader := io.LimitReader(in.File, in.MaxSizeBytes+1)

	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, err
	}

	if int64(len(data)) > in.MaxSizeBytes {
		return nil, ErrFileTooLarge
	}

	allFilesSize, err := uc.repositories.FileRepository.GetFilesSizeByUser(ctx, userEntity.ID)
	if err != nil {
		logging.GetLogger(ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}

	if (allFilesSize + int64(len(data))) > in.StorageMaxSize {
		return nil, ErrFileSystemIsFull
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
	extAllowed, allowed := uc.GetAllowedMimeTypes()[mimeType]
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
	fileExt = strings.TrimPrefix(fileExt, ".")

	safeName := filepath.Base(in.OriginalFilename)
	if strings.Contains(safeName, "..") {
		return nil, ErrFileNotSafeFilename
	}

	newFilename, err := fileService.GenerateNewFileName(fileExt)
	if err != nil {
		return nil, err
	}

	maxFileID, err := uc.repositories.FileRepository.GetLastID(ctx)
	if err != nil {
		logging.GetLogger(ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}

	middleFilePath := filepath.Join(fileService.GetMiddlePathByFileId(maxFileID+1), newFilename)
	fullFilePath := filepath.Join(in.SavePath, middleFilePath)

	saveDto := &dto.SaveFile{
		File:      bytes.NewReader(data),
		SavePath:  fullFilePath,
		SizeBytes: int64(len(data)),
	}

	saveErr := uc.repositories.StorageRepository.Save(ctx, saveDto)
	if saveErr != nil {
		return nil, ErrFileSave
	}

	fileHash, err := fileService.GenerateFileHash()
	if err != nil {
		return nil, err
	}

	fileEntity := &entity.File{
		UserID:           userEntity.ID,
		OriginalFilename: in.OriginalFilename,
		FilePath:         middleFilePath,
		Ext:              fileExt,
		Size:             len(data),
		Hash:             fileHash,
		CreatedAt:        time.Now().UTC(),
	}

	_, err = uc.repositories.FileRepository.Create(ctx, fileEntity)
	if err != nil {
		logging.GetLogger(ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}

	return fileEntity, nil
}

func (uc *fileUseCase) GetFileByHash(ctx context.Context, in dto.GetFileByHash) (*dto.FileResponse, error) {
	fileEntity, err := uc.repositories.FileRepository.GetByHash(ctx, in.Hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrFileNotFound
		}
		logging.GetLogger(ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}

	fullPath := filepath.Join(in.SavePath, fileEntity.FilePath)
	fileReader, err := uc.repositories.StorageRepository.GetFile(ctx, fullPath)
	if err != nil {
		logging.GetLogger(ctx).Error(err)
		return nil, err
	}

	fileResponse := &dto.FileResponse{
		File:             fileReader,
		OriginalFilename: fileEntity.OriginalFilename,
	}
	return fileResponse, nil
}

func (uc *fileUseCase) DeleteByID(ctx context.Context, fileID int, generalPath string) error {
	fileEntity, err := uc.repositories.FileRepository.GetByID(ctx, fileID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrFileNotFound
		}
		logging.GetLogger(ctx).Error(err)
		return postgres.ErrUnexpectedDBError
	}

	fullPath := filepath.Join(generalPath, fileEntity.FilePath)
	err = uc.repositories.StorageRepository.Delete(ctx, fullPath)
	if err != nil {
		logging.GetLogger(ctx).Error(err)
		return err
	}

	err = uc.repositories.FileRepository.DeleteByID(ctx, fileID)
	if err != nil {
		logging.GetLogger(ctx).Error(err)
		return err
	}

	return nil
}

func (uc *fileUseCase) CleanUnused(ctx context.Context, generalPath string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ch, err := uc.repositories.FileRepository.GetUnusedFileIDs(ctx)
	if err != nil {
		return err
	}

	for id := range ch {
		err := uc.DeleteByID(ctx, id, generalPath)
		if err != nil {
			return err
		}
	}
	return nil
}
