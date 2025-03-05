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
}

func NewRepositories(ctx context.Context, db *pgxpool.Pool) *Repositories {
	return &Repositories{
		UserRepository:         NewUserRepository(ctx, db),
		NoteRepository:         NewNoteRepository(ctx, db),
		NoteCategoryRepository: NewNoteCategoryRepository(ctx, db),
		BlockIPRepository:      NewBlockIpRepository(ctx, db),
	}
}
