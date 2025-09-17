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
	Save(ctx context.Context, in *dto.SaveFile) error
	GetFile(ctx context.Context, filePath string) (io.Reader, error)
	Delete(ctx context.Context, filePath string) error
	DeleteAll(ctx context.Context, filePaths []string) error
}

type localStorageRepository struct {
}
type s3StorageRepository struct {
	minio      *minio.Client
	bucketName string
}

func NewLocalStorageRepository() FileStorageRepository {
	return &localStorageRepository{}
}
func NewS3StorageRepository(minio *minio.Client, bucketName string) FileStorageRepository {
	return &s3StorageRepository{
		minio:      minio,
		bucketName: bucketName,
	}
}

func (r *localStorageRepository) Save(ctx context.Context, in *dto.SaveFile) error {
	dir := filepath.Dir(in.SavePath)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		logging.GetLogger(ctx).Errorf("failed to create directories: %v", err)
		return ErrFileSave
	}

	out, err := os.Create(in.SavePath)
	if err != nil {
		logging.GetLogger(ctx).Error(err)
		return ErrFileSave
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			logging.GetLogger(ctx).Errorf("error closing file: %v", err)
		}
	}(out)

	_, err = io.Copy(out, in.File)
	if err != nil {
		logging.GetLogger(ctx).Error(err)
		return ErrFileSave
	}
	return nil
}

func (r *localStorageRepository) GetFile(ctx context.Context, filePath string) (io.Reader, error) {
	file, err := os.Open(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrFileNotFoundInFilesystem
		}
		return nil, err
	}
	return file, nil
}

func (r *localStorageRepository) Delete(ctx context.Context, filePath string) error {
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

func (r *localStorageRepository) DeleteAll(ctx context.Context, filePaths []string) error {
	for _, filePath := range filePaths {
		err := r.Delete(ctx, filePath)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *s3StorageRepository) Save(ctx context.Context, in *dto.SaveFile) error {
	_, err := r.minio.PutObject(
		ctx,
		r.bucketName,
		in.SavePath,
		in.File,
		in.SizeBytes,
		minio.PutObjectOptions{ContentType: "application/octet-stream"},
	)
	if err != nil {
		logging.GetLogger(ctx).Error(err)
		return err
	}
	return nil
}

func (r *s3StorageRepository) GetFile(ctx context.Context, filePath string) (io.Reader, error) {
	object, err := r.minio.GetObject(ctx, r.bucketName, filePath, minio.GetObjectOptions{})
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

func (r *s3StorageRepository) Delete(ctx context.Context, filePath string) error {
	err := r.minio.RemoveObject(ctx, r.bucketName, filePath, minio.RemoveObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (r *s3StorageRepository) DeleteAll(ctx context.Context, filePaths []string) error {
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

	errorCh := r.minio.RemoveObjects(ctx, r.bucketName, objectCh, minio.RemoveObjectsOptions{})

	for err := range errorCh {
		logging.GetLogger(ctx).Errorf("Ошибка удаления объекта %s: %v", err.ObjectName, err.Err)
	}

	return nil
}
