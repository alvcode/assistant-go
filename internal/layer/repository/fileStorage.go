package repository

import (
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/logging"
	"context"
	"errors"
	"github.com/minio/minio-go/v7"
	"io"
	"os"
)

var (
	ErrFileSave = errors.New("unable to save file")
)

type FileStorageRepository interface {
	Save(in *dto.SaveFile) error
	GetFile(filePath string) (io.Reader, error)
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
	out, err := os.Create(in.SavePath)
	if err != nil {
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
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			logging.GetLogger(r.ctx).Error(err)
		}
	}(file)
	return file, nil
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

	return object, nil
}
