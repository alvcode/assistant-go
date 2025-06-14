package repository

import (
	"assistant-go/internal/config"
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
)

type DBExecutor interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Repositories struct {
	UserRepository         UserRepository
	NoteRepository         NoteRepository
	NoteCategoryRepository NoteCategoryRepository
	BlockIPRepository      BlockIPRepository
	BlockEventRepository   BlockEventRepository
	RateLimiterRepository  RateLimiterRepository
	FileRepository         FileRepository
	StorageRepository      FileStorageRepository
	FileNoteLinkRepository FileNoteLinkRepository
	TransactionRepository  TransactionRepository
	DriveStructRepository  DriveStructRepository
	DriveFileRepository    DriveFileRepository
}

func NewRepositories(ctx context.Context, cfg *config.Config, db *pgxpool.Pool, minio *minio.Client) *Repositories {
	var storageInterface FileStorageRepository
	if cfg.UploadPlace == config.FileUploadS3Place {
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
		RateLimiterRepository:  NewRateLimiterRepository(ctx, db),
		FileRepository:         NewFileRepository(ctx, db),
		StorageRepository:      storageInterface,
		FileNoteLinkRepository: NewFileNoteLinkRepository(ctx, db),
		TransactionRepository:  &transactionRepository{ctx: ctx, db: db},
		DriveStructRepository:  NewDriveStructRepository(ctx, db),
		DriveFileRepository:    NewDriveFileRepository(ctx, db),
	}
}

type TransactionRepository interface {
	GetTransaction() (pgx.Tx, error)
}

type transactionRepository struct {
	ctx context.Context
	db  *pgxpool.Pool
}

func (r *transactionRepository) GetTransaction() (pgx.Tx, error) {
	tx, err := r.db.BeginTx(r.ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	return tx, nil
}
