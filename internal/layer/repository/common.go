package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repositories struct {
	UserRepository         UserRepository
	NoteRepository         NoteRepository
	NoteCategoryRepository NoteCategoryRepository
	BlockIPRepository      BlockIPRepository
	BlockEventRepository   BlockEventRepository
	FileRepository         FileRepository
}

func NewRepositories(ctx context.Context, db *pgxpool.Pool) *Repositories {
	return &Repositories{
		UserRepository:         NewUserRepository(ctx, db),
		NoteRepository:         NewNoteRepository(ctx, db),
		NoteCategoryRepository: NewNoteCategoryRepository(ctx, db),
		BlockIPRepository:      NewBlockIpRepository(ctx, db),
		BlockEventRepository:   NewBlockEventRepository(ctx, db),
		FileRepository:         NewFileRepository(ctx, db),
	}
}
