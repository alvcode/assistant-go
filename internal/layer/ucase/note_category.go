package ucase

import (
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/layer/entity"
	"assistant-go/internal/layer/repository"
	service "assistant-go/internal/layer/service/note_category"
	"assistant-go/internal/logging"
	"assistant-go/internal/storage/postgres"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"reflect"
)

var (
	ErrCategoryParentIdNotFound = errors.New("parent category not found")
	ErrCategoryNotFound         = errors.New("category not found")
	ErrCategoryHasNotes         = errors.New("category has notes")
)

type NoteCategoryUseCase interface {
	Create(in dto.NoteCategoryCreate, userEntity *entity.User) (*entity.NoteCategory, error)
	FindAll(userId int) ([]*entity.NoteCategory, error)
	Delete(userId int, catId int) error
	Update(in dto.NoteCategoryUpdate, userID int) (*entity.NoteCategory, error)
	PositionUp(in dto.RequiredID, userID int) error
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

func (uc *noteCategoryUseCase) Create(in dto.NoteCategoryCreate, userEntity *entity.User) (*entity.NoteCategory, error) {
	noteCategoryEntity := entity.NoteCategory{
		UserId:   userEntity.ID,
		Name:     in.Name,
		ParentId: in.ParentId,
	}

	if in.ParentId != nil {
		_, err := uc.repositories.NoteCategoryRepository.FindByIDAndUser(uc.ctx, userEntity.ID, *in.ParentId)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrCategoryParentIdNotFound
			}
			logging.GetLogger(uc.ctx).Error(err)
			return nil, postgres.ErrUnexpectedDBError
		}
	}

	positionService := service.NewNoteCategory().PositionService(uc.ctx, uc.repositories)
	newPosition, err := positionService.CalculateForNew(userEntity.ID, in.ParentId)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}

	noteCategoryEntity.Position = newPosition

	data, err := uc.repositories.NoteCategoryRepository.Create(uc.ctx, noteCategoryEntity)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}
	return data, nil
}

func (uc *noteCategoryUseCase) FindAll(userId int) ([]*entity.NoteCategory, error) {
	data, err := uc.repositories.NoteCategoryRepository.FindAll(uc.ctx, userId)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}
	return data, nil
}

func (uc *noteCategoryUseCase) Delete(userId int, catId int) error {
	categories, err := uc.repositories.NoteCategoryRepository.FindByIDAndUserWithChildren(uc.ctx, userId, catId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrCategoryNotFound
		}
		logging.GetLogger(uc.ctx).Error(err)
		return postgres.ErrUnexpectedDBError
	}

	catIds := make([]int, 0)
	for _, cat := range categories {
		catIds = append(catIds, cat.ID)
	}
	if len(catIds) == 0 {
		return ErrCategoryNotFound
	}

	checkExists, err := uc.repositories.NoteRepository.CheckExistsByCategoryIDs(catIds)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			logging.GetLogger(uc.ctx).Error(err)
			return postgres.ErrUnexpectedDBError
		}
	}
	if checkExists == true {
		return ErrCategoryHasNotes
	}

	err = uc.repositories.NoteCategoryRepository.DeleteByIds(uc.ctx, catIds)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return postgres.ErrUnexpectedDBError
	}
	return nil
}

func (uc *noteCategoryUseCase) Update(in dto.NoteCategoryUpdate, userID int) (*entity.NoteCategory, error) {
	noteCategoryEntity := &entity.NoteCategory{
		ID:       in.ID,
		UserId:   userID,
		Name:     in.Name,
		ParentId: in.ParentID,
	}

	currentCategory, err := uc.repositories.NoteCategoryRepository.FindByIDAndUser(uc.ctx, userID, in.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}
		logging.GetLogger(uc.ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}
	noteCategoryEntity.Position = currentCategory.Position

	if in.ParentID != nil {
		_, err = uc.repositories.NoteCategoryRepository.FindByIDAndUser(uc.ctx, userID, *in.ParentID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrCategoryNotFound
			}
			logging.GetLogger(uc.ctx).Error(err)
			return nil, postgres.ErrUnexpectedDBError
		}
	}

	if !reflect.DeepEqual(in.ParentID, currentCategory.ParentId) {
		positionService := service.NewNoteCategory().PositionService(uc.ctx, uc.repositories)
		newPosition, err := positionService.CalculateForNew(userID, in.ParentID)
		if err != nil {
			logging.GetLogger(uc.ctx).Error(err)
			return nil, postgres.ErrUnexpectedDBError
		}
		noteCategoryEntity.Position = newPosition
	}

	err = uc.repositories.NoteCategoryRepository.Update(uc.ctx, noteCategoryEntity)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}
	return noteCategoryEntity, nil
}

func (uc *noteCategoryUseCase) PositionUp(in dto.RequiredID, userID int) error {
	_, err := uc.repositories.NoteCategoryRepository.FindByIDAndUser(uc.ctx, userID, in.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrCategoryNotFound
		}
		logging.GetLogger(uc.ctx).Error(err)
		return postgres.ErrUnexpectedDBError
	}

	positionService := service.NewNoteCategory().PositionService(uc.ctx, uc.repositories)
	err = positionService.PositionUp(userID, in.ID)
	if err != nil {
		return err
	}
	return nil
}
