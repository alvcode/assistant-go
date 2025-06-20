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
	"path/filepath"
	"strings"
	"time"
)

const (
	typeDirectory = 0
	typeFile      = 1
)

var (
	ErrDriveFileTooLarge        = errors.New("drive file too large")
	ErrDriveFileSystemIsFull    = errors.New("drive file system is full")
	ErrDriveFileNotSafeFilename = errors.New("drive file not safe filename")
	ErrDriveFileSave            = errors.New("drive unable to save file")
)

var (
	ErrDriveDirectoryExists  = errors.New("directory exists")
	ErrDriveParentIdNotFound = errors.New("drive parent id does not exist")
)

type DriveUseCase interface {
	CreateDirectory(dto *dto.DriveCreateDirectory, user *entity.User) ([]*entity.DriveStruct, error)
	GetTree(parentID *int, user *entity.User) ([]*entity.DriveStruct, error)
	UploadFile(in dto.DriveUploadFile, user *entity.User) ([]*entity.DriveStruct, error)
	Delete(structID int, user *entity.User) error
}

type driveUseCase struct {
	ctx          context.Context
	repositories *repository.Repositories
}

func NewDriveUseCase(ctx context.Context, repositories *repository.Repositories) DriveUseCase {
	return &driveUseCase{
		ctx:          ctx,
		repositories: repositories,
	}
}

func (uc *driveUseCase) GetTree(parentID *int, user *entity.User) ([]*entity.DriveStruct, error) {
	list, err := uc.repositories.DriveStructRepository.ListByUserID(user.ID, parentID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			logging.GetLogger(uc.ctx).Error(err)
			return nil, postgres.ErrUnexpectedDBError
		}
	}
	return list, nil
}

func (uc *driveUseCase) CreateDirectory(dto *dto.DriveCreateDirectory, user *entity.User) ([]*entity.DriveStruct, error) {
	if dto.ParentID != nil {
		parentStruct, err := uc.repositories.DriveStructRepository.GetByID(*dto.ParentID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrDriveParentIdNotFound
			}
		}
		if parentStruct.UserID != user.ID {
			return nil, ErrDriveDirectoryExists
		}
	}
	_, err := uc.repositories.DriveStructRepository.FindRow(user.ID, dto.Name, typeDirectory, dto.ParentID)

	if err == nil {
		return nil, ErrDriveDirectoryExists
	}
	createEntity := &entity.DriveStruct{
		UserID:    user.ID,
		Name:      dto.Name,
		Type:      typeDirectory,
		ParentID:  dto.ParentID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	_, err = uc.repositories.DriveStructRepository.Create(createEntity)
	if err != nil {
		return nil, err
	}

	treeList, err := uc.GetTree(dto.ParentID, user)
	if err != nil {
		return nil, err
	}
	return treeList, nil
}

func (uc *driveUseCase) UploadFile(in dto.DriveUploadFile, user *entity.User) ([]*entity.DriveStruct, error) {
	fileService := service.NewFile().FileService()
	limitedReader := io.LimitReader(in.File, in.MaxSizeBytes+1)

	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, err
	}

	if int64(len(data)) > in.MaxSizeBytes {
		return nil, ErrDriveFileTooLarge
	}

	allStorageSize, err := uc.repositories.DriveFileRepository.GetStorageSize(user.ID)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}

	if (allStorageSize + int64(len(data))) > in.StorageMaxSizePerUser {
		return nil, ErrDriveFileSystemIsFull
	}

	if seeker, ok := in.File.(io.Seeker); ok {
		_, err := seeker.Seek(0, io.SeekStart)
		if err != nil {
			return nil, ErrFileResettingPointer
		}
	} else {
		return nil, ErrFileUnableToSeek
	}

	fileExt := strings.ToLower(filepath.Ext(in.OriginalFilename))
	fileExt = strings.TrimPrefix(fileExt, ".")

	safeName := filepath.Base(in.OriginalFilename)
	if strings.Contains(safeName, "..") {
		return nil, ErrDriveFileNotSafeFilename
	}

	newFilename, err := fileService.GenerateNewFileName(fileExt)
	if err != nil {
		return nil, err
	}

	maxFileID, err := uc.repositories.DriveFileRepository.GetLastID()
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}

	middleFilePath := filepath.Join(fileService.GetMiddlePathByFileId(maxFileID+1), newFilename)
	fullFilePath := filepath.Join(in.SavePath, middleFilePath)

	saveDto := &dto.SaveFile{
		File:      bytes.NewReader(data),
		SavePath:  fullFilePath,
		SizeBytes: int64(len(data)),
	}

	saveErr := uc.repositories.StorageRepository.Save(saveDto)
	if saveErr != nil {
		return nil, ErrDriveFileSave
	}

	// сохраняем 2 записи в БД
	driveStruct := &entity.DriveStruct{
		UserID:    user.ID,
		Name:      in.OriginalFilename,
		Type:      typeFile,
		ParentID:  in.ParentID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	driveStruct, err = uc.repositories.DriveStructRepository.Create(driveStruct)
	if err != nil {
		return nil, err
	}

	driveFile := &entity.DriveFile{
		DriveStructID: driveStruct.ID,
		Path:          middleFilePath,
		Ext:           fileExt,
		Size:          len(data),
		CreatedAt:     time.Now().UTC(),
	}

	_, err = uc.repositories.DriveFileRepository.Create(driveFile)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}

	treeList, err := uc.GetTree(in.ParentID, user)
	if err != nil {
		return nil, err
	}
	return treeList, nil
}

func (uc *driveUseCase) Delete(structID int, user *entity.User) error {

}
