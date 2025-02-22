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
)

type NoteCategoryUseCase interface {
	Create(in dto.NoteCategoryCreate, userEntity *entity.User, lang string) (*entity.NoteCategory, error)
	FindAll(userId int, lang string) ([]*entity.NoteCategory, error)
	Delete(userId int, catId int, lang string) error
	DeleteByUserId(userId int, lang string) error
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
	in dto.NoteCategoryCreate,
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

func (uc *noteCategoryUseCase) FindAll(userId int, lang string) ([]*entity.NoteCategory, error) {
	data, err := uc.noteCategoryRepository.FindAll(userId)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, errors.New(locale.T(lang, "unexpected_database_error"))
	}
	return data, nil
}

func (uc *noteCategoryUseCase) Delete(userId int, catId int, lang string) error {
	user, err := uc.noteCategoryRepository.FindByIDAndUser(userId, catId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New(locale.T(lang, "category_not_found"))
		}
	}
	if user.UserId != userId {
		return errors.New(locale.T(lang, "category_not_found"))
	}

	// TODO: тут должна быть проверка на наличие заметок внутри категории и ошибка, т.к тогда они потеряются

	err = uc.noteCategoryRepository.DeleteById(catId)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return errors.New(locale.T(lang, "unexpected_database_error"))
	}
	return nil
}

func (uc *noteCategoryUseCase) DeleteByUserId(userId int, lang string) error {
	_, err := uc.FindAll(userId, lang)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return errors.New(locale.T(lang, "unexpected_database_error"))
	}
	err = uc.noteCategoryRepository.DeleteByUserId(userId)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return errors.New(locale.T(lang, "unexpected_database_error"))
	}
	return nil
}
