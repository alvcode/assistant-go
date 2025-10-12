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
	UserRepository           UserRepository
	NoteRepository           NoteRepository
	NoteCategoryRepository   NoteCategoryRepository
	BlockIPRepository        BlockIPRepository
	BlockEventRepository     BlockEventRepository
	RateLimiterRepository    RateLimiterRepository
	FileRepository           FileRepository
	StorageRepository        FileStorageRepository
	FileNoteLinkRepository   FileNoteLinkRepository
	TransactionRepository    TransactionRepository
	DriveStructRepository    DriveStructRepository
	DriveFileRepository      DriveFileRepository
	DriveFileChunkRepository DriveFileChunkRepository
}

func NewRepositories(ctx context.Context, cfg *config.Config, db *pgxpool.Pool, minio *minio.Client) *Repositories {
	var storageInterface FileStorageRepository
	if cfg.UploadPlace == config.FileUploadS3Place {
		storageInterface = NewS3StorageRepository(minio, cfg.S3.BucketName)
	} else {
		storageInterface = NewLocalStorageRepository()
	}
	return &Repositories{
		UserRepository:           NewUserRepository(db),
		NoteRepository:           NewNoteRepository(db),
		NoteCategoryRepository:   NewNoteCategoryRepository(db),
		BlockIPRepository:        NewBlockIpRepository(db),
		BlockEventRepository:     NewBlockEventRepository(db),
		RateLimiterRepository:    NewRateLimiterRepository(db),
		FileRepository:           NewFileRepository(db),
		StorageRepository:        storageInterface,
		FileNoteLinkRepository:   NewFileNoteLinkRepository(db),
		TransactionRepository:    &transactionRepository{db: db},
		DriveStructRepository:    NewDriveStructRepository(db),
		DriveFileRepository:      NewDriveFileRepository(db),
		DriveFileChunkRepository: NewDriveFileChunkRepository(db),
	}
}

type TransactionRepository interface {
	GetTransaction(ctx context.Context) (pgx.Tx, error)
}

type transactionRepository struct {
	db *pgxpool.Pool
}

func (r *transactionRepository) GetTransaction(ctx context.Context) (pgx.Tx, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func WithTransaction(ctx context.Context, tr TransactionRepository, fn func(tx pgx.Tx) error) error {
	tx, err := tr.GetTransaction(ctx)
	if err != nil {
		return err
	}

	err = fn(tx)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

func WithTransactionResult[T any](
	ctx context.Context,
	tr TransactionRepository,
	fn func(tx pgx.Tx) (T, error),
) (T, error) {
	var zero T

	tx, err := tr.GetTransaction(ctx)
	if err != nil {
		return zero, err
	}

	result, err := fn(tx)
	if err != nil {
		_ = tx.Rollback(ctx)
		return zero, err
	}

	return result, tx.Commit(ctx)
}
