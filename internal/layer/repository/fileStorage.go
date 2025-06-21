package repository

import (
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/logging"
	"context"
	"errors"
	"github.com/minio/minio-go/v7"
	"io"
	"os"
	"path/filepath"
)

var (
	ErrFileSave                 = errors.New("unable to save file")
	ErrFileNotFoundInFilesystem = errors.New("file not found in filesystem")
)

type FileStorageRepository interface {
	Save(in *dto.SaveFile) error
	GetFile(filePath string) (io.Reader, error)
	Delete(filePath string) error
	DeleteAll(filePaths []string) error
}

type localStorageRepository struct {
	ctx context.Context
}
type s3StorageRepository struct {
	ctx        context.Context
	minio      *minio.Client
	bucketName string
}

func NewLocalStorageRepository(ctx context.Context) FileStorageRepository {
	return &localStorageRepository{
		ctx: ctx,
	}
}
func NewS3StorageRepository(ctx context.Context, minio *minio.Client, bucketName string) FileStorageRepository {
	return &s3StorageRepository{
		ctx:        ctx,
		minio:      minio,
		bucketName: bucketName,
	}
}

func (r *localStorageRepository) Save(in *dto.SaveFile) error {
	dir := filepath.Dir(in.SavePath)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		logging.GetLogger(r.ctx).Errorf("failed to create directories: %v", err)
		return ErrFileSave
	}

	out, err := os.Create(in.SavePath)
	if err != nil {
		logging.GetLogger(r.ctx).Error(err)
		return ErrFileSave
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			logging.GetLogger(r.ctx).Errorf("error closing file: %v", err)
		}
	}(out)

	_, err = io.Copy(out, in.File)
	if err != nil {
		logging.GetLogger(r.ctx).Error(err)
		return ErrFileSave
	}
	return nil
}

func (r *localStorageRepository) GetFile(filePath string) (io.Reader, error) {
	file, err := os.Open(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrFileNotFoundInFilesystem
		}
		return nil, err
	}
	return file, nil
}

func (r *localStorageRepository) Delete(filePath string) error {
	_, err := os.Stat(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrFileNotFoundInFilesystem
		}
		return err
	}
	err = os.Remove(filePath)
	if err != nil {
		return err
	}
	return nil
}

func (r *localStorageRepository) DeleteAll(filePaths []string) error {
	for _, filePath := range filePaths {
		err := r.Delete(filePath)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *s3StorageRepository) Save(in *dto.SaveFile) error {
	_, err := r.minio.PutObject(
		r.ctx,
		r.bucketName,
		in.SavePath,
		in.File,
		in.SizeBytes,
		minio.PutObjectOptions{ContentType: "application/octet-stream"},
	)
	if err != nil {
		logging.GetLogger(r.ctx).Error(err)
		return err
	}
	return nil
}

func (r *s3StorageRepository) GetFile(filePath string) (io.Reader, error) {
	object, err := r.minio.GetObject(r.ctx, r.bucketName, filePath, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	_, err = object.Stat()
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return nil, ErrFileNotFoundInFilesystem
		}
		return nil, err
	}

	return object, nil
}

func (r *s3StorageRepository) Delete(filePath string) error {
	err := r.minio.RemoveObject(r.ctx, r.bucketName, filePath, minio.RemoveObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (r *s3StorageRepository) DeleteAll(filePaths []string) error {
	objectCh := make(chan minio.ObjectInfo)

	// Горутина для отправки объектов в канал
	go func() {
		defer close(objectCh)
		for _, key := range filePaths {
			objectCh <- minio.ObjectInfo{
				Key: key,
			}
		}
	}()

	errorCh := r.minio.RemoveObjects(r.ctx, r.bucketName, objectCh, minio.RemoveObjectsOptions{})

	for err := range errorCh {
		logging.GetLogger(r.ctx).Errorf("Ошибка удаления объекта %s: %v", err.ObjectName, err.Err)
	}

	return nil
}
