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
	Update(catID int, catName string, catParentID *int, userID int, lang string) (*entity.NoteCategory, error)
}

type noteCategoryUseCase struct {
	ctx          context.Context
	repositories *repository.Repositories
}

func NewNoteCategoryUseCase(ctx context.Context, repositories *repository.Repositories) NoteCategoryUseCase {
	return &noteCategoryUseCase{
		ctx:          ctx,
		repositories: repositories,
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
		_, err := uc.repositories.NoteCategoryRepository.FindByIDAndUser(userEntity.ID, *in.ParentId)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, errors.New(locale.T(lang, "parent_id_of_the_category_not_found"))
			}
			logging.GetLogger(uc.ctx).Error(err)
			return nil, errors.New(locale.T(lang, "unexpected_database_error"))
		}
	}

	data, err := uc.repositories.NoteCategoryRepository.Create(noteCategoryEntity)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, errors.New(locale.T(lang, "unexpected_database_error"))
	}
	return data, nil
}

func (uc *noteCategoryUseCase) FindAll(userId int, lang string) ([]*entity.NoteCategory, error) {
	data, err := uc.repositories.NoteCategoryRepository.FindAll(userId)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, errors.New(locale.T(lang, "unexpected_database_error"))
	}
	return data, nil
}

func (uc *noteCategoryUseCase) Delete(userId int, catId int, lang string) error {
	_, err := uc.repositories.NoteCategoryRepository.FindByIDAndUser(userId, catId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New(locale.T(lang, "category_not_found"))
		}
		logging.GetLogger(uc.ctx).Error(err)
		return errors.New(locale.T(lang, "unexpected_database_error"))
	}

	// TODO: тут должна быть проверка на наличие заметок внутри категории и ошибка, т.к тогда они потеряются

	err = uc.repositories.NoteCategoryRepository.DeleteById(catId)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return errors.New(locale.T(lang, "unexpected_database_error"))
	}
	return nil
}

func (uc *noteCategoryUseCase) Update(catID int, catName string, catParentID *int, userID int, lang string) (*entity.NoteCategory, error) {
	noteCategoryEntity := entity.NoteCategory{
		ID:       catID,
		UserId:   userID,
		Name:     catName,
		ParentId: catParentID,
	}

	if catParentID != nil {
		_, err := uc.repositories.NoteCategoryRepository.FindByIDAndUser(userID, *catParentID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, errors.New(locale.T(lang, "parent_id_of_the_category_not_found"))
			}
			logging.GetLogger(uc.ctx).Error(err)
			return nil, errors.New(locale.T(lang, "unexpected_database_error"))
		}
	}

	data, err := uc.repositories.NoteCategoryRepository.Update(noteCategoryEntity)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, errors.New(locale.T(lang, "unexpected_database_error"))
	}
	return data, nil

}
