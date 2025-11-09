package ucase

import (
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/layer/entity"
	"assistant-go/internal/layer/repository"
	service "assistant-go/internal/layer/service/file"
	"assistant-go/internal/logging"
	"assistant-go/internal/storage/postgres"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	typeDirectory = 0
	typeFile      = 1
)

var (
	ErrDriveFileTooLarge                    = errors.New("drive file too large")
	ErrDriveFileTooLargeUseChunks           = errors.New("drive file too large use chunks")
	ErrDriveFileSystemIsFull                = errors.New("drive file system is full")
	ErrDriveFileNotSafeFilename             = errors.New("drive file not safe filename")
	ErrDriveFileSave                        = errors.New("drive unable to save file")
	ErrDriveDirectoryExists                 = errors.New("directory exists")
	ErrDriveFilenameExists                  = errors.New("drive filename exists")
	ErrDriveParentIdNotFound                = errors.New("drive parent id does not exist")
	ErrDriveStructNotFound                  = errors.New("drive struct not found")
	ErrDriveRelocatableStructureNotFound    = errors.New("drive relocatable structure not found")
	ErrDriveMovingIntoOneself               = errors.New("drive moving into oneself")
	ErrDriveParentRefOfTheRelocatableStruct = errors.New("drive parent ref of the relocatable struct")
	ErrDriveUnavailableForChunks            = errors.New("drive unavailable for chunks")
)

type DriveUseCase interface {
	CreateDirectory(dto *dto.DriveCreateDirectory, user *entity.User) ([]*dto.DriveTree, error)
	GetTree(parentID *int, user *entity.User) ([]*dto.DriveTree, error)
	UploadFile(in dto.DriveUploadFile, user *entity.User) ([]*dto.DriveTree, error)
	Delete(structID int, savePath string, user *entity.User) error
	GetFile(structID int, savePath string, user *entity.User) (*dto.FileResponse, error)
	Rename(structID int, newName string, user *entity.User) error
	Space(user *entity.User, totalSpace int64) (*dto.DriveSpace, error)
	RenMov(user *entity.User, in dto.DriveRenMov) error
	ChunkPrepare(user *entity.User, in dto.DriveChunkPrepareIn) (*dto.DriveChunkPrepareResponse, error)
	ChunkUpload(user *entity.User, in dto.DriveUploadChunk) error
	ChunkEnd(structID int) error
	ChunksInfo(structID int) (*dto.DriveChunksInfo, error)
	GetChunkBytes(structID int, chunkNumber int, savePath string, user *entity.User) (*dto.FileResponse, error)
	UpdateFileHash(structID int, hash string, user *entity.User) error
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
	list, err := uc.repositories.DriveStructRepository.TreeByUserID(uc.ctx, user.ID, parentID)
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
		err := uc.checkParentOwner(*dto.ParentID, user.ID)
		if err != nil {
			return nil, err
		}
	}
	_, err := uc.repositories.DriveStructRepository.FindRow(uc.ctx, user.ID, dto.Name, typeDirectory, dto.ParentID)

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
	_, err = uc.repositories.DriveStructRepository.Create(uc.ctx, createEntity)
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

	size, err := uc.getFileSize(in.File, in.MaxSizeBytes)
	if err != nil {
		return nil, err
	}

	if size > 64<<20 {
		return nil, ErrDriveFileTooLargeUseChunks
	}

	if size > in.MaxSizeBytes {
		return nil, ErrDriveFileTooLarge
	}

	allStorageSize, err := uc.repositories.DriveFileRepository.GetStorageSize(uc.ctx, user.ID)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}

	if (allStorageSize + size) > in.StorageMaxSizePerUser {
		return nil, ErrDriveFileSystemIsFull
	}

	fileExt := strings.ToLower(filepath.Ext(in.OriginalFilename))
	fileExt = strings.TrimPrefix(fileExt, ".")

	safeName := filepath.Base(in.OriginalFilename)
	if strings.Contains(safeName, "..") {
		return nil, ErrDriveFileNotSafeFilename
	}

	_, err = uc.repositories.DriveStructRepository.FindRow(uc.ctx, user.ID, in.OriginalFilename, typeFile, in.ParentID)

	if err == nil {
		return nil, ErrDriveFilenameExists
	}

	newFilename, err := fileService.GenerateNewFileName(fileExt)
	if err != nil {
		return nil, err
	}

	maxFileID, err := uc.repositories.DriveFileRepository.GetLastID(uc.ctx)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}

	middleFilePath := filepath.Join(fileService.GetMiddlePathByFileId(maxFileID+1), newFilename)
	fullFilePath := filepath.Join(in.SavePath, middleFilePath)

	limitedReader := io.LimitReader(in.File, in.MaxSizeBytes+1)
	saveDto := &dto.SaveFile{
		File:      limitedReader,
		SavePath:  fullFilePath,
		SizeBytes: size,
	}

	saveErr := uc.repositories.StorageRepository.Save(uc.ctx, saveDto)
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

	driveStruct, err = uc.repositories.DriveStructRepository.Create(uc.ctx, driveStruct)
	if err != nil {
		return nil, err
	}

	driveFile := &entity.DriveFile{
		DriveStructID: driveStruct.ID,
		Path:          &middleFilePath,
		Ext:           fileExt,
		Size:          size,
		CreatedAt:     time.Now().UTC(),
		IsChunk:       false,
		SHA256:        in.SHA256,
	}

	_, err = uc.repositories.DriveFileRepository.Create(uc.ctx, driveFile)
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
	driveStruct, err := uc.repositories.DriveStructRepository.GetByID(uc.ctx, structID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrDriveStructNotFound
		}
	}
	if driveStruct.UserID != user.ID {
		return ErrDriveStructNotFound
	}

	deleteChunkList, err := uc.repositories.DriveFileChunkRepository.GetAllRecursive(uc.ctx, structID, user.ID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
	}

	deleteFileList, err := uc.repositories.DriveFileRepository.GetAllRecursive(uc.ctx, structID, user.ID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
	}

	var keys []string
	if len(deleteChunkList) > 0 {
		for _, fileChunk := range deleteChunkList {
			keys = append(keys, filepath.Join(savePath, fileChunk.Path))
		}
	}
	if len(deleteFileList) > 0 {
		for _, file := range deleteFileList {
			if !file.IsChunk {
				keys = append(keys, filepath.Join(savePath, *file.Path))
			}
		}
	}

	if len(keys) > 0 {
		_ = uc.repositories.StorageRepository.DeleteAll(uc.ctx, keys)
	}

	// удаление записей из БД из трех таблиц (через cascade fk)
	err = uc.repositories.DriveStructRepository.DeleteRecursive(uc.ctx, user.ID, structID)
	if err != nil {
		return err
	}

	return nil
}

func (uc *driveUseCase) GetFile(structID int, savePath string, user *entity.User) (*dto.FileResponse, error) {
	driveStruct, err := uc.repositories.DriveStructRepository.GetByID(uc.ctx, structID)
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

	driveFile, err := uc.repositories.DriveFileRepository.GetByStructID(uc.ctx, driveStruct.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrFileNotFound
		}
	}
	if driveFile.IsChunk {
		return nil, ErrDriveUnavailableForChunks
	}

	fullPath := filepath.Join(savePath, *driveFile.Path)
	fileReader, err := uc.repositories.StorageRepository.GetFile(uc.ctx, fullPath)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, err
	}

	fileResponse := &dto.FileResponse{
		File:             fileReader,
		OriginalFilename: driveStruct.Name,
		SizeBytes:        driveFile.Size,
	}
	return fileResponse, nil
}

func (uc *driveUseCase) Rename(structID int, newName string, user *entity.User) error {
	driveStruct, err := uc.repositories.DriveStructRepository.GetByID(uc.ctx, structID)
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
	err = uc.repositories.DriveStructRepository.Update(uc.ctx, driveStruct)
	if err != nil {
		return err
	}

	return nil
}

func (uc *driveUseCase) Space(user *entity.User, totalSpace int64) (*dto.DriveSpace, error) {
	result := &dto.DriveSpace{}
	result.Total = totalSpace

	usedSpace, err := uc.repositories.DriveFileRepository.GetStorageSize(uc.ctx, user.ID)
	if err != nil {
		return nil, err
	}
	result.Used = usedSpace

	return result, nil
}

func (uc *driveUseCase) RenMov(user *entity.User, in dto.DriveRenMov) error {
	if in.ParentID != nil {
		parentStruct, err := uc.repositories.DriveStructRepository.GetByID(uc.ctx, *in.ParentID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrDriveParentIdNotFound
			}
			return err
		}

		if parentStruct.UserID != user.ID {
			return ErrDriveParentIdNotFound
		}
		if parentStruct.Type != typeDirectory {
			return ErrDriveParentIdNotFound
		}
		if parentStruct.ParentID != nil {
			for _, structID := range in.StructIDs {
				if structID == *parentStruct.ParentID {
					return ErrDriveParentRefOfTheRelocatableStruct
				}
			}
		}
	}

	var batches [][]int
	batchSize := 100
	for i, structID := range in.StructIDs {
		if in.ParentID != nil && *in.ParentID == structID {
			return ErrDriveMovingIntoOneself
		}

		if i%batchSize == 0 {
			batches = append(batches, []int{})
		}
		batches[len(batches)-1] = append(batches[len(batches)-1], structID)
	}

	for _, batch := range batches {
		structCount, err := uc.repositories.DriveStructRepository.StructCountByUserAndIDs(uc.ctx, user.ID, batch)
		if err != nil {
			return err
		}
		if structCount != len(batch) {
			return ErrDriveRelocatableStructureNotFound
		}
	}

	err := repository.WithTransaction(uc.ctx, uc.repositories.TransactionRepository, func(tx pgx.Tx) error {
		driveStructRepoTx := repository.NewDriveStructRepository(tx)

		for _, batch := range batches {
			err := driveStructRepoTx.MassUpdateParentID(uc.ctx, in.ParentID, batch)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return err
	}
	return nil
}

func (uc *driveUseCase) ChunkPrepare(user *entity.User, in dto.DriveChunkPrepareIn) (*dto.DriveChunkPrepareResponse, error) {
	if in.FullSize > in.MaxSizeBytes {
		return nil, ErrDriveFileTooLarge
	}

	allStorageSize, err := uc.repositories.DriveFileRepository.GetStorageSize(uc.ctx, user.ID)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}

	if (allStorageSize + in.FullSize) > in.StorageMaxSizePerUser {
		return nil, ErrDriveFileSystemIsFull
	}

	if in.ParentID != nil {
		err := uc.checkParentOwner(*in.ParentID, user.ID)
		if err != nil {
			return nil, err
		}
	}

	fileExt := strings.ToLower(filepath.Ext(in.Filename))
	fileExt = strings.TrimPrefix(fileExt, ".")

	safeName := filepath.Base(in.Filename)
	if strings.Contains(safeName, "..") {
		return nil, ErrDriveFileNotSafeFilename
	}

	_, err = uc.repositories.DriveStructRepository.FindRow(uc.ctx, user.ID, in.Filename, typeFile, in.ParentID)

	if err == nil {
		return nil, ErrDriveFilenameExists
	}

	driveStruct := &entity.DriveStruct{
		UserID:    user.ID,
		Name:      in.Filename,
		Type:      typeFile,
		ParentID:  in.ParentID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	driveStructResult, err := repository.WithTransactionResult(
		uc.ctx,
		uc.repositories.TransactionRepository,
		func(tx pgx.Tx) (*entity.DriveStruct, error) {
			driveStructRepo := repository.NewDriveStructRepository(tx)

			driveStruct, err = driveStructRepo.Create(uc.ctx, driveStruct)
			if err != nil {
				return nil, err
			}

			driveFile := &entity.DriveFile{
				DriveStructID: driveStruct.ID,
				Path:          nil,
				Ext:           fileExt,
				Size:          0,
				CreatedAt:     time.Now().UTC(),
				IsChunk:       true,
				SHA256:        in.DriveChunkPrepare.SHA256,
			}

			driveFileRepo := repository.NewDriveFileRepository(tx)

			driveFile, err = driveFileRepo.Create(uc.ctx, driveFile)
			if err != nil {
				logging.GetLogger(uc.ctx).Error(err)
				return nil, postgres.ErrUnexpectedDBError
			}

			return driveStruct, nil
		})

	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, err
	}

	return &dto.DriveChunkPrepareResponse{StructID: driveStructResult.ID}, nil
}

func (uc *driveUseCase) ChunkUpload(user *entity.User, in dto.DriveUploadChunk) error {
	fileService := service.NewFile().FileService()

	size, err := uc.getFileSize(in.File, in.MaxSizeBytes)
	if err != nil {
		return err
	}

	if size > 64<<20 {
		return ErrDriveFileTooLargeUseChunks
	}

	fileEntity, err := uc.repositories.DriveFileRepository.GetByStructID(uc.ctx, in.StructID)
	if err != nil {
		return err
	}

	checkFileOwner, err := uc.repositories.DriveFileRepository.CheckFileOwner(uc.ctx, fileEntity.ID, user.ID)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return err
	}
	if !checkFileOwner {
		return ErrDriveStructNotFound
	}

	chunksSize, err := uc.repositories.DriveFileChunkRepository.GetChunksSize(uc.ctx, fileEntity.ID)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return err
	}

	if chunksSize+size > in.MaxSizeBytes {
		return ErrDriveFileTooLarge
	}

	allStorageSize, err := uc.repositories.DriveFileRepository.GetStorageSize(uc.ctx, user.ID)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return postgres.ErrUnexpectedDBError
	}

	if (allStorageSize + chunksSize + size) > in.StorageMaxSizePerUser {
		return ErrDriveFileSystemIsFull
	}

	newFilename, err := fileService.GenerateNewFileName(fmt.Sprintf("%s_%d", "part", in.ChunkNumber))
	if err != nil {
		return err
	}

	middleFilePath := filepath.Join(fileService.GetMiddlePathByFileId(fileEntity.ID), newFilename)
	fullFilePath := filepath.Join(in.SavePath, middleFilePath)

	limitedReader := io.LimitReader(in.File, 65<<20)
	saveDto := &dto.SaveFile{
		File:      limitedReader,
		SavePath:  fullFilePath,
		SizeBytes: size,
	}

	saveErr := uc.repositories.StorageRepository.Save(uc.ctx, saveDto)
	if saveErr != nil {
		return ErrDriveFileSave
	}

	driveFileChunk := &entity.DriveFileChunk{
		DriveFileID: fileEntity.ID,
		Path:        middleFilePath,
		Size:        size,
		ChunkNumber: in.ChunkNumber,
	}

	_, err = uc.repositories.DriveFileChunkRepository.Create(uc.ctx, driveFileChunk)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return postgres.ErrUnexpectedDBError
	}

	return nil
}

func (uc *driveUseCase) ChunkEnd(structID int) error {
	fileEntity, err := uc.repositories.DriveFileRepository.GetByStructID(uc.ctx, structID)
	if err != nil {
		return err
	}

	chunksSize, err := uc.repositories.DriveFileChunkRepository.GetChunksSize(uc.ctx, fileEntity.ID)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return err
	}

	err = uc.repositories.DriveFileRepository.UpdateSize(uc.ctx, fileEntity.ID, chunksSize)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return err
	}
	return nil
}

func (uc *driveUseCase) ChunksInfo(structID int) (*dto.DriveChunksInfo, error) {
	fileEntity, err := uc.repositories.DriveFileRepository.GetByStructID(uc.ctx, structID)
	if err != nil {
		if errors.Is(pgx.ErrNoRows, err) {
			return nil, ErrDriveStructNotFound
		}
		logging.GetLogger(uc.ctx).Error(err)
		return nil, err
	}

	chunksInfo, err := uc.repositories.DriveFileChunkRepository.GetChunksInfo(uc.ctx, fileEntity.ID)
	if err != nil {
		if !errors.Is(pgx.ErrNoRows, err) {
			logging.GetLogger(uc.ctx).Error(err)
		}
		return nil, err
	}
	return chunksInfo, nil
}

func (uc *driveUseCase) GetChunkBytes(structID int, chunkNumber int, savePath string, user *entity.User) (*dto.FileResponse, error) {
	driveStruct, err := uc.repositories.DriveStructRepository.GetByID(uc.ctx, structID)
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

	driveFile, err := uc.repositories.DriveFileRepository.GetByStructID(uc.ctx, driveStruct.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrFileNotFound
		}
		return nil, err
	}

	driveFileChunk, err := uc.repositories.DriveFileChunkRepository.GetByFileIDAndNumber(uc.ctx, driveFile.ID, chunkNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrFileNotFound
		}
		return nil, err
	}

	fullPath := filepath.Join(savePath, driveFileChunk.Path)
	fileReader, err := uc.repositories.StorageRepository.GetFile(uc.ctx, fullPath)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, err
	}

	fileResponse := &dto.FileResponse{
		File:             fileReader,
		OriginalFilename: driveStruct.Name,
		SizeBytes:        driveFileChunk.Size,
	}
	return fileResponse, nil
}

func (uc *driveUseCase) UpdateFileHash(structID int, hash string, user *entity.User) error {
	driveStruct, err := uc.repositories.DriveStructRepository.GetByID(uc.ctx, structID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrFileNotFound
		}
		return err
	}
	if driveStruct.UserID != user.ID {
		return ErrFileNotFound
	}
	if driveStruct.Type != typeFile {
		return ErrFileNotFound
	}

	driveFile, err := uc.repositories.DriveFileRepository.GetByStructID(uc.ctx, driveStruct.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrFileNotFound
		}
		return err
	}

	err = uc.repositories.DriveFileRepository.UpdateHash(uc.ctx, driveFile.ID, hash)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return err
	}
	return nil
}

func (uc *driveUseCase) getFileSize(file multipart.File, maxSize int64) (int64, error) {
	var size int64
	// Если файл поддерживает Stat():
	if statter, ok := file.(interface{ Stat() (os.FileInfo, error) }); ok {
		if fi, err := statter.Stat(); err == nil {
			size = fi.Size()
		}
	}

	// Если не получилось через Stat() — считаем вручную через LimitedReader:
	if size == 0 {
		lr := io.LimitReader(file, maxSize+1)
		n, err := io.Copy(io.Discard, lr)
		if err != nil {
			return 0, err
		}
		size = n
	}

	if seeker, ok := file.(io.Seeker); ok {
		_, err := seeker.Seek(0, io.SeekStart)
		if err != nil {
			return 0, ErrFileResettingPointer
		}
	} else {
		return 0, ErrFileUnableToSeek
	}

	return size, nil
}

func (uc *driveUseCase) checkParentOwner(parentID int, userID int) error {
	parentStruct, err := uc.repositories.DriveStructRepository.GetByID(uc.ctx, parentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrDriveParentIdNotFound
		}
		return err
	}
	if parentStruct.UserID != userID {
		return ErrDriveDirectoryExists
	}
	return nil
}
