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
	ErrDriveDirectoryExists     = errors.New("directory exists")
	ErrDriveFilenameExists      = errors.New("drive filename exists")
	ErrDriveParentIdNotFound    = errors.New("drive parent id does not exist")
	ErrDriveStructNotFound      = errors.New("drive struct not found")
)

type DriveUseCase interface {
	CreateDirectory(dto *dto.DriveCreateDirectory, user *entity.User) ([]*dto.DriveTree, error)
	GetTree(parentID *int, user *entity.User) ([]*dto.DriveTree, error)
	UploadFile(in dto.DriveUploadFile, user *entity.User) ([]*dto.DriveTree, error)
	Delete(structID int, savePath string, user *entity.User) error
	GetFile(structID int, savePath string, user *entity.User) (*dto.FileResponse, error)
	Rename(structID int, newName string, user *entity.User) error
	Space(user *entity.User, totalSpace int64) (*dto.DriveSpace, error)
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

func (uc *driveUseCase) GetTree(parentID *int, user *entity.User) ([]*dto.DriveTree, error) {
	list, err := uc.repositories.DriveStructRepository.TreeByUserID(user.ID, parentID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			logging.GetLogger(uc.ctx).Error(err)
			return nil, postgres.ErrUnexpectedDBError
		}
	}
	return list, nil
}

func (uc *driveUseCase) CreateDirectory(dto *dto.DriveCreateDirectory, user *entity.User) ([]*dto.DriveTree, error) {
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

func (uc *driveUseCase) UploadFile(in dto.DriveUploadFile, user *entity.User) ([]*dto.DriveTree, error) {
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

	_, err = uc.repositories.DriveStructRepository.FindRow(user.ID, in.OriginalFilename, typeFile, in.ParentID)

	if err == nil {
		return nil, ErrDriveFilenameExists
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

func (uc *driveUseCase) Delete(structID int, savePath string, user *entity.User) error {
	driveStruct, err := uc.repositories.DriveStructRepository.GetByID(structID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrDriveStructNotFound
		}
	}
	if driveStruct.UserID != user.ID {
		return ErrDriveStructNotFound
	}

	existsFiles := true
	deleteFileList, err := uc.repositories.DriveFileRepository.GetAllRecursive(structID, user.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			existsFiles = false
		} else {
			return err
		}
	}

	if existsFiles {
		var keys []string
		for _, file := range deleteFileList {
			keys = append(keys, filepath.Join(savePath, file.Path))
		}

		if len(keys) > 0 {
			_ = uc.repositories.StorageRepository.DeleteAll(keys)
		}
	}

	// удаление записей из БД из двух таблиц
	err = uc.repositories.DriveStructRepository.DeleteRecursive(user.ID, structID)
	if err != nil {
		return err
	}

	return nil
}

func (uc *driveUseCase) GetFile(structID int, savePath string, user *entity.User) (*dto.FileResponse, error) {
	driveStruct, err := uc.repositories.DriveStructRepository.GetByID(structID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrFileNotFound
		}
		return nil, err
	}
	if driveStruct.UserID != user.ID {
		return nil, ErrFileNotFound
	}
	if driveStruct.Type != typeFile {
		return nil, ErrFileNotFound
	}

	driveFile, err := uc.repositories.DriveFileRepository.GetByStructID(driveStruct.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrFileNotFound
		}
	}

	fullPath := filepath.Join(savePath, driveFile.Path)
	fileReader, err := uc.repositories.StorageRepository.GetFile(fullPath)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, err
	}

	fileResponse := &dto.FileResponse{
		File:             fileReader,
		OriginalFilename: driveStruct.Name,
	}
	return fileResponse, nil
}

func (uc *driveUseCase) Rename(structID int, newName string, user *entity.User) error {
	driveStruct, err := uc.repositories.DriveStructRepository.GetByID(structID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrFileNotFound
		}
		return err
	}
	if driveStruct.UserID != user.ID {
		return ErrFileNotFound
	}

	driveStruct.Name = newName
	err = uc.repositories.DriveStructRepository.Update(driveStruct)
	if err != nil {
		return err
	}

	return nil
}

func (uc *driveUseCase) Space(user *entity.User, totalSpace int64) (*dto.DriveSpace, error) {
	result := &dto.DriveSpace{}
	result.Total = totalSpace

	usedSpace, err := uc.repositories.DriveFileRepository.GetStorageSize(user.ID)
	if err != nil {
		return nil, err
	}
	result.Used = usedSpace

	return result, nil
}
