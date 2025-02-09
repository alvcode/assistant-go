package useCase

import (
	dtoNoteCategory "assistant-go/internal/layer/dto/noteCategory"
	"assistant-go/internal/layer/entity"
	"assistant-go/internal/layer/repository"
	"assistant-go/internal/locale"
	"assistant-go/internal/logging"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
)

type NoteCategoryUseCase interface {
	Create(in dtoNoteCategory.Create, userEntity *entity.User, lang string) (*entity.NoteCategory, error)
	FindAll(userId uint32, lang string) ([]*entity.NoteCategory, error)
}

type noteCategoryUseCase struct {
	ctx                    context.Context
	noteCategoryRepository repository.NoteCategoryRepository
}

func NewNoteCategoryUseCase(ctx context.Context, noteCategoryRepository repository.NoteCategoryRepository) NoteCategoryUseCase {
	return &noteCategoryUseCase{
		ctx:                    ctx,
		noteCategoryRepository: noteCategoryRepository,
	}
}

func (uc *noteCategoryUseCase) Create(
	in dtoNoteCategory.Create,
	userEntity *entity.User,
	lang string,
) (*entity.NoteCategory, error) {
	noteCategoryEntity := entity.NoteCategory{
		UserId:   userEntity.ID,
		Name:     in.Name,
		ParentId: in.ParentId,
	}

	if in.ParentId != nil {
		_, err := uc.noteCategoryRepository.FindByIDAndUser(userEntity.ID, *in.ParentId)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, errors.New(locale.T(lang, "parent_id_of_the_category_not_found"))
			}
			logging.GetLogger(uc.ctx).Error(err)
			return nil, errors.New(locale.T(lang, "unexpected_database_error"))
		}
	}

	data, err := uc.noteCategoryRepository.Create(noteCategoryEntity)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, errors.New(locale.T(lang, "unexpected_database_error"))
	}
	return data, nil
}

func (uc *noteCategoryUseCase) FindAll(userId uint32, lang string) ([]*entity.NoteCategory, error) {
	data, err := uc.noteCategoryRepository.FindAll(userId)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, errors.New(locale.T(lang, "unexpected_database_error"))
	}
	return data, nil
}
