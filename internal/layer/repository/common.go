package repository

import (
	"assistant-go/internal/config"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
)

type Repositories struct {
	UserRepository         UserRepository
	NoteRepository         NoteRepository
	NoteCategoryRepository NoteCategoryRepository
	BlockIPRepository      BlockIPRepository
	BlockEventRepository   BlockEventRepository
	FileRepository         FileRepository
	StorageRepository      FileStorageRepository
}

func NewRepositories(ctx context.Context, cfg *config.Config, db *pgxpool.Pool, minio *minio.Client) *Repositories {
	var storageInterface FileStorageRepository
	if cfg.File.UploadPlace == config.FileUploadS3Place {
		storageInterface = NewS3StorageRepository(ctx, minio, cfg.S3.BucketName)
	} else {
		storageInterface = NewLocalStorageRepository(ctx)
	}
	return &Repositories{
		UserRepository:         NewUserRepository(ctx, db),
		NoteRepository:         NewNoteRepository(ctx, db),
		NoteCategoryRepository: NewNoteCategoryRepository(ctx, db),
		BlockIPRepository:      NewBlockIpRepository(ctx, db),
		BlockEventRepository:   NewBlockEventRepository(ctx, db),
		FileRepository:         NewFileRepository(ctx, db),
		StorageRepository:      storageInterface,
	}
}
