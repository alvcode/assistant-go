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
	Update(in dto.NoteCategoryUpdate, userID int, lang string) (*entity.NoteCategory, error)
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
	categories, err := uc.repositories.NoteCategoryRepository.FindByIDAndUserWithChildren(userId, catId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New(locale.T(lang, "category_not_found"))
		}
		logging.GetLogger(uc.ctx).Error(err)
		return errors.New(locale.T(lang, "unexpected_database_error"))
	}

	catIds := make([]int, 0)
	for _, cat := range categories {
		catIds = append(catIds, cat.ID)
	}
	if len(catIds) == 0 {
		return errors.New(locale.T(lang, "category_not_found"))
	}

	checkExists, err := uc.repositories.NoteRepository.CheckExistsByCategoryIDs(catIds)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New(locale.T(lang, "category_not_found"))
		}
		logging.GetLogger(uc.ctx).Error(err)
		return errors.New(locale.T(lang, "unexpected_database_error"))
	}
	if checkExists == true {
		return errors.New(locale.T(lang, "category_has_notes"))
	}

	err = uc.repositories.NoteCategoryRepository.DeleteByIds(catIds)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return errors.New(locale.T(lang, "unexpected_database_error"))
	}
	return nil
}

func (uc *noteCategoryUseCase) Update(in dto.NoteCategoryUpdate, userID int, lang string) (*entity.NoteCategory, error) {
	noteCategoryEntity := &entity.NoteCategory{
		ID:       in.ID,
		UserId:   userID,
		Name:     in.Name,
		ParentId: in.ParentID,
	}

	if in.ParentID != nil {
		_, err := uc.repositories.NoteCategoryRepository.FindByIDAndUser(userID, *in.ParentID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, errors.New(locale.T(lang, "category_not_found"))
			}
			logging.GetLogger(uc.ctx).Error(err)
			return nil, errors.New(locale.T(lang, "unexpected_database_error"))
		}
	}

	err := uc.repositories.NoteCategoryRepository.Update(noteCategoryEntity)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, errors.New(locale.T(lang, "unexpected_database_error"))
	}
	return noteCategoryEntity, nil

}
