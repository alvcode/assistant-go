package repository

import (
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/logging"
	"context"
	"errors"
	"fmt"
	"github.com/minio/minio-go/v7"
	"io"
	"os"
)

var (
	ErrFileUnableToSave = errors.New("unable to save file")
	ErrFileSave         = errors.New("unable to save file")
)

type FileStorageRepository interface {
	Save(in *dto.SaveFile) error
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

func (ur *localStorageRepository) Save(in *dto.SaveFile) error {
	out, err := os.Create(in.SavePath)
	if err != nil {
		return ErrFileUnableToSave
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			logging.GetLogger(ur.ctx).Errorf("error closing file: %v", err)
		}
	}(out)

	_, err = io.Copy(out, in.File)
	if err != nil {
		return ErrFileSave
	}
	return nil
}

func (ur *s3StorageRepository) Save(in *dto.SaveFile) error {
	uploadInfo, err := ur.minio.PutObject(
		ur.ctx,
		ur.bucketName,
		in.SavePath,
		in.File,
		in.SizeBytes,
		minio.PutObjectOptions{ContentType: "application/octet-stream"},
	)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(uploadInfo)
	return nil
}
