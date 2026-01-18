package ucase

import (
	"assistant-go/internal/layer/entity"
	"assistant-go/internal/layer/repository"
	"assistant-go/internal/logging"
	"assistant-go/internal/storage/postgres"
	"context"
	"fmt"
)

type NoteShareUseCase interface {
	Create(ctx context.Context, noteID int, userEntity *entity.User) (*entity.NoteShare, error)
}

type noteShareUseCase struct {
	repositories repository.Repositories
}

func NewNoteShareUseCase(repositories *repository.Repositories) NoteShareUseCase {
	return &noteShareUseCase{
		repositories: *repositories,
	}
}

func (uc *noteShareUseCase) Create(ctx context.Context, noteID int, userEntity *entity.User) (*entity.NoteShare, error) {
	noteBelongsUser, err := uc.repositories.NoteRepository.BelongsToUser(ctx, noteID, userEntity.ID)
	if err != nil {
		logging.GetLogger(ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}

	if !noteBelongsUser {
		return nil, ErrNoteNotFound
	}

	fmt.Println(noteID, userEntity.ID)

	return nil, nil
}
