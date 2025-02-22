package ucase

import (
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/layer/entity"
	"assistant-go/internal/layer/repository"
	"assistant-go/internal/locale"
	"assistant-go/internal/logging"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"time"
)

type NoteUseCase interface {
	Create(in dto.NoteCreate, userEntity *entity.User, lang string) (*entity.Note, error)
}

type noteUseCase struct {
	ctx          context.Context
	repositories repository.Repositories
}

func NewNoteUseCase(ctx context.Context, repositories *repository.Repositories) NoteUseCase {
	return &noteUseCase{
		ctx:          ctx,
		repositories: *repositories,
	}
}

func (uc *noteUseCase) Create(
	in dto.NoteCreate,
	userEntity *entity.User,
	lang string,
) (*entity.Note, error) {
	_, err := uc.repositories.NoteCategoryRepository.FindByIDAndUser(userEntity.ID, in.CategoryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(locale.T(lang, "category_not_found"))
		}
		logging.GetLogger(uc.ctx).Error(err)
		return nil, errors.New(locale.T(lang, "unexpected_database_error"))
	}

	timeNow := time.Now().UTC()

	noteEntity := entity.Note{
		CategoryID: in.CategoryID,
		NoteBlocks: in.NoteBlocks,
		CreatedAt:  timeNow,
		UpdatedAt:  timeNow,
	}

	data, err := uc.repositories.NoteRepository.Create(noteEntity)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, errors.New(locale.T(lang, "unexpected_database_error"))
	}
	return data, nil
}
