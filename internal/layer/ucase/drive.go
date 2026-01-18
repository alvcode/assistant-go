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
	ErrDriveEncrypting                      = errors.New("error encrypting file")
	ErrDriveDecrypting                      = errors.New("error decrypting file")
)

type DriveUseCase interface {
	CreateDirectory(ctx context.Context, dto *dto.DriveCreateDirectory, user *entity.User) ([]*dto.DriveTree, error)
	GetTree(ctx context.Context, parentID *int, user *entity.User) ([]*dto.DriveTree, error)
	UploadFile(ctx context.Context, in dto.DriveUploadFile, user *entity.User) ([]*dto.DriveTree, error)
	Delete(ctx context.Context, structID int, savePath string, user *entity.User) error
	GetFile(ctx context.Context, in *dto.GetFile, user *entity.User) (*dto.FileResponse, error)
	Rename(ctx context.Context, structID int, newName string, user *entity.User) error
	Space(ctx context.Context, user *entity.User, totalSpace int64) (*dto.DriveSpace, error)
	RenMov(ctx context.Context, user *entity.User, in dto.DriveRenMov) error
	ChunkPrepare(ctx context.Context, user *entity.User, in dto.DriveChunkPrepareIn) (*dto.DriveChunkPrepareResponse, error)
	ChunkUpload(ctx context.Context, user *entity.User, in dto.DriveUploadChunk) error
	ChunkEnd(ctx context.Context, structID int) error
	ChunksInfo(ctx context.Context, structID int) (*dto.DriveChunksInfo, error)
	GetChunkBytes(ctx context.Context, in *dto.GetChunk, user *entity.User) (*dto.FileResponse, error)
	UpdateFileHash(ctx context.Context, structID int, hash string, user *entity.User) error
}

type driveUseCase struct {
	repositories *repository.Repositories
}

func NewDriveUseCase(repositories *repository.Repositories) DriveUseCase {
	return &driveUseCase{
		repositories: repositories,
	}
}

func (uc *driveUseCase) GetTree(ctx context.Context, parentID *int, user *entity.User) ([]*dto.DriveTree, error) {
	list, err := uc.repositories.DriveStructRepository.TreeByUserID(ctx, user.ID, parentID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			logging.GetLogger(ctx).Error(err)
			return nil, postgres.ErrUnexpectedDBError
		}
	}
	return list, nil
}

func (uc *driveUseCase) CreateDirectory(ctx context.Context, dto *dto.DriveCreateDirectory, user *entity.User) ([]*dto.DriveTree, error) {
	if dto.ParentID != nil {
		err := uc.checkParentOwner(ctx, *dto.ParentID, user.ID)
		if err != nil {
			return nil, err
		}
	}
	_, err := uc.repositories.DriveStructRepository.FindRow(ctx, user.ID, dto.Name, typeDirectory, dto.ParentID)

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
	_, err = uc.repositories.DriveStructRepository.Create(ctx, createEntity)
	if err != nil {
		return nil, err
	}

	treeList, err := uc.GetTree(ctx, dto.ParentID, user)
	if err != nil {
		return nil, err
	}
	return treeList, nil
}

func (uc *driveUseCase) UploadFile(ctx context.Context, in dto.DriveUploadFile, user *entity.User) ([]*dto.DriveTree, error) {
	fileService := service.NewFile().FileService()

	if in.UseEncryption {
		var encryptErr error
		in.File, encryptErr = fileService.EncryptFile(in.File, in.EncryptionKey)
		if encryptErr != nil {
			logging.GetLogger(ctx).Error(fmt.Errorf("%w: %w", ErrDriveEncrypting, encryptErr))
			return nil, ErrDriveEncrypting
		}
	}

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

	allStorageSize, err := uc.repositories.DriveFileRepository.GetStorageSize(ctx, user.ID)
	if err != nil {
		logging.GetLogger(ctx).Error(err)
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

	_, err = uc.repositories.DriveStructRepository.FindRow(ctx, user.ID, in.OriginalFilename, typeFile, in.ParentID)

	if err == nil {
		return nil, ErrDriveFilenameExists
	}

	newFilename, err := fileService.GenerateNewFileName(fileExt)
	if err != nil {
		return nil, err
	}

	maxFileID, err := uc.repositories.DriveFileRepository.GetLastID(ctx)
	if err != nil {
		logging.GetLogger(ctx).Error(err)
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

	saveErr := uc.repositories.StorageRepository.Save(ctx, saveDto)
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

	driveStruct, err = uc.repositories.DriveStructRepository.Create(ctx, driveStruct)
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

	_, err = uc.repositories.DriveFileRepository.Create(ctx, driveFile)
	if err != nil {
		logging.GetLogger(ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}

	treeList, err := uc.GetTree(ctx, in.ParentID, user)
	if err != nil {
		return nil, err
	}

	return treeList, nil
}

func (uc *driveUseCase) Delete(ctx context.Context, structID int, savePath string, user *entity.User) error {
	driveStruct, err := uc.repositories.DriveStructRepository.GetByID(ctx, structID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrDriveStructNotFound
		}
	}
	if driveStruct.UserID != user.ID {
		return ErrDriveStructNotFound
	}

	deleteChunkList, err := uc.repositories.DriveFileChunkRepository.GetAllRecursive(ctx, structID, user.ID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
	}

	deleteFileList, err := uc.repositories.DriveFileRepository.GetAllRecursive(ctx, structID, user.ID)
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
		_ = uc.repositories.StorageRepository.DeleteAll(ctx, keys)
	}

	// удаление записей из БД из трех таблиц (через cascade fk)
	err = uc.repositories.DriveStructRepository.DeleteRecursive(ctx, user.ID, structID)
	if err != nil {
		return err
	}

	return nil
}

func (uc *driveUseCase) GetFile(ctx context.Context, in *dto.GetFile, user *entity.User) (*dto.FileResponse, error) {
	driveStruct, err := uc.repositories.DriveStructRepository.GetByID(ctx, in.StructID)
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

	driveFile, err := uc.repositories.DriveFileRepository.GetByStructID(ctx, driveStruct.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrFileNotFound
		}
	}
	if driveFile.IsChunk {
		return nil, ErrDriveUnavailableForChunks
	}

	fullPath := filepath.Join(in.SavePath, *driveFile.Path)
	fileReader, err := uc.repositories.StorageRepository.GetFile(ctx, fullPath)
	if err != nil {
		logging.GetLogger(ctx).Error(err)
		return nil, err
	}

	realSize := driveFile.Size
	if in.UseEncryption {
		fileService := service.NewFile().FileService()
		fileReader, err = fileService.DecryptFile(fileReader, in.EncryptionKey)
		if err != nil {
			logging.GetLogger(ctx).Error(fmt.Errorf("%w: %w", ErrDriveDecrypting, err))
			return nil, ErrDriveDecrypting
		}

		realSize, err = uc.getFileSizeInReader(fileReader, in.MaxSizeBytes)
		if err != nil {
			return nil, err
		}
	}

	fileResponse := &dto.FileResponse{
		File:             fileReader,
		OriginalFilename: driveStruct.Name,
		SizeBytes:        realSize,
	}
	return fileResponse, nil
}

func (uc *driveUseCase) Rename(ctx context.Context, structID int, newName string, user *entity.User) error {
	driveStruct, err := uc.repositories.DriveStructRepository.GetByID(ctx, structID)
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
	err = uc.repositories.DriveStructRepository.Update(ctx, driveStruct)
	if err != nil {
		return err
	}

	return nil
}

func (uc *driveUseCase) Space(ctx context.Context, user *entity.User, totalSpace int64) (*dto.DriveSpace, error) {
	result := &dto.DriveSpace{}
	result.Total = totalSpace

	usedSpace, err := uc.repositories.DriveFileRepository.GetStorageSize(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	result.Used = usedSpace

	return result, nil
}

func (uc *driveUseCase) RenMov(ctx context.Context, user *entity.User, in dto.DriveRenMov) error {
	if in.ParentID != nil {
		parentStruct, err := uc.repositories.DriveStructRepository.GetByID(ctx, *in.ParentID)
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
		structCount, err := uc.repositories.DriveStructRepository.StructCountByUserAndIDs(ctx, user.ID, batch)
		if err != nil {
			return err
		}
		if structCount != len(batch) {
			return ErrDriveRelocatableStructureNotFound
		}
	}

	err := repository.WithTransaction(ctx, uc.repositories.TransactionRepository, func(tx pgx.Tx) error {
		driveStructRepoTx := repository.NewDriveStructRepository(tx)

		for _, batch := range batches {
			err := driveStructRepoTx.MassUpdateParentID(ctx, in.ParentID, batch)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		logging.GetLogger(ctx).Error(err)
		return err
	}
	return nil
}

func (uc *driveUseCase) ChunkPrepare(ctx context.Context, user *entity.User, in dto.DriveChunkPrepareIn) (*dto.DriveChunkPrepareResponse, error) {
	if in.FullSize > in.MaxSizeBytes {
		return nil, ErrDriveFileTooLarge
	}

	allStorageSize, err := uc.repositories.DriveFileRepository.GetStorageSize(ctx, user.ID)
	if err != nil {
		logging.GetLogger(ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}

	if (allStorageSize + in.FullSize) > in.StorageMaxSizePerUser {
		return nil, ErrDriveFileSystemIsFull
	}

	if in.ParentID != nil {
		err := uc.checkParentOwner(ctx, *in.ParentID, user.ID)
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

	_, err = uc.repositories.DriveStructRepository.FindRow(ctx, user.ID, in.Filename, typeFile, in.ParentID)

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
		ctx,
		uc.repositories.TransactionRepository,
		func(tx pgx.Tx) (*entity.DriveStruct, error) {
			driveStructRepo := repository.NewDriveStructRepository(tx)

			driveStruct, err = driveStructRepo.Create(ctx, driveStruct)
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

			driveFile, err = driveFileRepo.Create(ctx, driveFile)
			if err != nil {
				logging.GetLogger(ctx).Error(err)
				return nil, postgres.ErrUnexpectedDBError
			}

			return driveStruct, nil
		})

	if err != nil {
		logging.GetLogger(ctx).Error(err)
		return nil, err
	}

	return &dto.DriveChunkPrepareResponse{StructID: driveStructResult.ID}, nil
}

func (uc *driveUseCase) ChunkUpload(ctx context.Context, user *entity.User, in dto.DriveUploadChunk) error {
	fileService := service.NewFile().FileService()

	if in.UseEncryption {
		var encryptErr error
		in.File, encryptErr = fileService.EncryptFile(in.File, in.EncryptionKey)
		if encryptErr != nil {
			logging.GetLogger(ctx).Error(fmt.Errorf("%w: %w", ErrDriveEncrypting, encryptErr))
			return ErrDriveEncrypting
		}
	}

	size, err := uc.getFileSize(in.File, in.MaxSizeBytes)
	if err != nil {
		return err
	}

	if size > 64<<20 {
		return ErrDriveFileTooLargeUseChunks
	}

	fileEntity, err := uc.repositories.DriveFileRepository.GetByStructID(ctx, in.StructID)
	if err != nil {
		return err
	}

	checkFileOwner, err := uc.repositories.DriveFileRepository.CheckFileOwner(ctx, fileEntity.ID, user.ID)
	if err != nil {
		logging.GetLogger(ctx).Error(err)
		return err
	}
	if !checkFileOwner {
		return ErrDriveStructNotFound
	}

	chunksSize, err := uc.repositories.DriveFileChunkRepository.GetChunksSize(ctx, fileEntity.ID)
	if err != nil {
		logging.GetLogger(ctx).Error(err)
		return err
	}

	if chunksSize+size > in.MaxSizeBytes {
		return ErrDriveFileTooLarge
	}

	allStorageSize, err := uc.repositories.DriveFileRepository.GetStorageSize(ctx, user.ID)
	if err != nil {
		logging.GetLogger(ctx).Error(err)
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

	saveErr := uc.repositories.StorageRepository.Save(ctx, saveDto)
	if saveErr != nil {
		return ErrDriveFileSave
	}

	driveFileChunk := &entity.DriveFileChunk{
		DriveFileID: fileEntity.ID,
		Path:        middleFilePath,
		Size:        size,
		ChunkNumber: in.ChunkNumber,
	}

	_, err = uc.repositories.DriveFileChunkRepository.Create(ctx, driveFileChunk)
	if err != nil {
		logging.GetLogger(ctx).Error(err)
		return postgres.ErrUnexpectedDBError
	}

	return nil
}

func (uc *driveUseCase) ChunkEnd(ctx context.Context, structID int) error {
	fileEntity, err := uc.repositories.DriveFileRepository.GetByStructID(ctx, structID)
	if err != nil {
		return err
	}

	chunksSize, err := uc.repositories.DriveFileChunkRepository.GetChunksSize(ctx, fileEntity.ID)
	if err != nil {
		logging.GetLogger(ctx).Error(err)
		return err
	}

	err = uc.repositories.DriveFileRepository.UpdateSize(ctx, fileEntity.ID, chunksSize)
	if err != nil {
		logging.GetLogger(ctx).Error(err)
		return err
	}
	return nil
}

func (uc *driveUseCase) ChunksInfo(ctx context.Context, structID int) (*dto.DriveChunksInfo, error) {
	fileEntity, err := uc.repositories.DriveFileRepository.GetByStructID(ctx, structID)
	if err != nil {
		if errors.Is(pgx.ErrNoRows, err) {
			return nil, ErrDriveStructNotFound
		}
		logging.GetLogger(ctx).Error(err)
		return nil, err
	}

	chunksInfo, err := uc.repositories.DriveFileChunkRepository.GetChunksInfo(ctx, fileEntity.ID)
	if err != nil {
		if !errors.Is(pgx.ErrNoRows, err) {
			logging.GetLogger(ctx).Error(err)
		}
		return nil, err
	}
	return chunksInfo, nil
}

func (uc *driveUseCase) GetChunkBytes(ctx context.Context, in *dto.GetChunk, user *entity.User) (*dto.FileResponse, error) {
	driveStruct, err := uc.repositories.DriveStructRepository.GetByID(ctx, in.StructID)
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

	driveFile, err := uc.repositories.DriveFileRepository.GetByStructID(ctx, driveStruct.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrFileNotFound
		}
		return nil, err
	}

	driveFileChunk, err := uc.repositories.DriveFileChunkRepository.GetByFileIDAndNumber(ctx, driveFile.ID, in.ChunkNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrFileNotFound
		}
		return nil, err
	}

	fullPath := filepath.Join(in.SavePath, driveFileChunk.Path)
	fileReader, err := uc.repositories.StorageRepository.GetFile(ctx, fullPath)
	if err != nil {
		logging.GetLogger(ctx).Error(err)
		return nil, err
	}

	realSize := driveFileChunk.Size
	if in.UseEncryption {
		fileService := service.NewFile().FileService()
		fileReader, err = fileService.DecryptFile(fileReader, in.EncryptionKey)
		if err != nil {
			logging.GetLogger(ctx).Error(fmt.Errorf("%w: %w", ErrDriveDecrypting, err))
			return nil, ErrDriveDecrypting
		}

		realSize, err = uc.getFileSizeInReader(fileReader, in.MaxSizeBytes)
		if err != nil {
			return nil, err
		}
	}

	fileResponse := &dto.FileResponse{
		File:             fileReader,
		OriginalFilename: driveStruct.Name,
		SizeBytes:        realSize,
	}
	return fileResponse, nil
}

func (uc *driveUseCase) UpdateFileHash(ctx context.Context, structID int, hash string, user *entity.User) error {
	driveStruct, err := uc.repositories.DriveStructRepository.GetByID(ctx, structID)
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

	driveFile, err := uc.repositories.DriveFileRepository.GetByStructID(ctx, driveStruct.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrFileNotFound
		}
		return err
	}

	err = uc.repositories.DriveFileRepository.UpdateHash(ctx, driveFile.ID, hash)
	if err != nil {
		logging.GetLogger(ctx).Error(err)
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

func (uc *driveUseCase) getFileSizeInReader(file io.Reader, maxSize int64) (int64, error) {
	var size int64

	lr := io.LimitReader(file, maxSize+1)
	n, err := io.Copy(io.Discard, lr)
	if err != nil {
		return 0, err
	}
	size = n

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

func (uc *driveUseCase) checkParentOwner(ctx context.Context, parentID int, userID int) error {
	parentStruct, err := uc.repositories.DriveStructRepository.GetByID(ctx, parentID)
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
