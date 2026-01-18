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
	Create(ctx context.Context, in dto.NoteCategoryCreate, userEntity *entity.User) (*entity.NoteCategory, error)
	FindAll(ctx context.Context, userId int) ([]*entity.NoteCategory, error)
	Delete(ctx context.Context, userId int, catId int) error
	Update(ctx context.Context, in dto.NoteCategoryUpdate, userID int) (*entity.NoteCategory, error)
	PositionUp(ctx context.Context, in dto.RequiredID, userID int) error
}

type noteCategoryUseCase struct {
	repositories *repository.Repositories
}

func NewNoteCategoryUseCase(repositories *repository.Repositories) NoteCategoryUseCase {
	return &noteCategoryUseCase{
		repositories: repositories,
	}
}

func (uc *noteCategoryUseCase) Create(ctx context.Context, in dto.NoteCategoryCreate, userEntity *entity.User) (*entity.NoteCategory, error) {
	noteCategoryEntity := entity.NoteCategory{
		UserId:   userEntity.ID,
		Name:     in.Name,
		ParentId: in.ParentId,
	}

	if in.ParentId != nil {
		_, err := uc.repositories.NoteCategoryRepository.FindByIDAndUser(ctx, userEntity.ID, *in.ParentId)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrCategoryParentIdNotFound
			}
			logging.GetLogger(ctx).Error(err)
			return nil, postgres.ErrUnexpectedDBError
		}
	}

	positionService := service.NewNoteCategory().PositionService(ctx, uc.repositories)
	newPosition, err := positionService.CalculateForNew(userEntity.ID, in.ParentId)
	if err != nil {
		logging.GetLogger(ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}

	noteCategoryEntity.Position = newPosition

	data, err := uc.repositories.NoteCategoryRepository.Create(ctx, noteCategoryEntity)
	if err != nil {
		logging.GetLogger(ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}
	return data, nil
}

func (uc *noteCategoryUseCase) FindAll(ctx context.Context, userId int) ([]*entity.NoteCategory, error) {
	data, err := uc.repositories.NoteCategoryRepository.FindAll(ctx, userId)
	if err != nil {
		logging.GetLogger(ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}
	return data, nil
}

func (uc *noteCategoryUseCase) Delete(ctx context.Context, userId int, catId int) error {
	categories, err := uc.repositories.NoteCategoryRepository.FindByIDAndUserWithChildren(ctx, userId, catId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrCategoryNotFound
		}
		logging.GetLogger(ctx).Error(err)
		return postgres.ErrUnexpectedDBError
	}

	catIds := make([]int, 0)
	for _, cat := range categories {
		catIds = append(catIds, cat.ID)
	}
	if len(catIds) == 0 {
		return ErrCategoryNotFound
	}

	checkExists, err := uc.repositories.NoteRepository.CheckExistsByCategoryIDs(ctx, catIds)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			logging.GetLogger(ctx).Error(err)
			return postgres.ErrUnexpectedDBError
		}
	}
	if checkExists == true {
		return ErrCategoryHasNotes
	}

	err = uc.repositories.NoteCategoryRepository.DeleteByIds(ctx, catIds)
	if err != nil {
		logging.GetLogger(ctx).Error(err)
		return postgres.ErrUnexpectedDBError
	}
	return nil
}

func (uc *noteCategoryUseCase) Update(ctx context.Context, in dto.NoteCategoryUpdate, userID int) (*entity.NoteCategory, error) {
	noteCategoryEntity := &entity.NoteCategory{
		ID:       in.ID,
		UserId:   userID,
		Name:     in.Name,
		ParentId: in.ParentID,
	}

	currentCategory, err := uc.repositories.NoteCategoryRepository.FindByIDAndUser(ctx, userID, in.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}
		logging.GetLogger(ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}
	noteCategoryEntity.Position = currentCategory.Position

	if in.ParentID != nil {
		_, err = uc.repositories.NoteCategoryRepository.FindByIDAndUser(ctx, userID, *in.ParentID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrCategoryNotFound
			}
			logging.GetLogger(ctx).Error(err)
			return nil, postgres.ErrUnexpectedDBError
		}
	}

	if !reflect.DeepEqual(in.ParentID, currentCategory.ParentId) {
		positionService := service.NewNoteCategory().PositionService(ctx, uc.repositories)
		newPosition, err := positionService.CalculateForNew(userID, in.ParentID)
		if err != nil {
			logging.GetLogger(ctx).Error(err)
			return nil, postgres.ErrUnexpectedDBError
		}
		noteCategoryEntity.Position = newPosition
	}

	err = uc.repositories.NoteCategoryRepository.Update(ctx, noteCategoryEntity)
	if err != nil {
		logging.GetLogger(ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}
	return noteCategoryEntity, nil
}

func (uc *noteCategoryUseCase) PositionUp(ctx context.Context, in dto.RequiredID, userID int) error {
	_, err := uc.repositories.NoteCategoryRepository.FindByIDAndUser(ctx, userID, in.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrCategoryNotFound
		}
		logging.GetLogger(ctx).Error(err)
		return postgres.ErrUnexpectedDBError
	}

	positionService := service.NewNoteCategory().PositionService(ctx, uc.repositories)
	err = positionService.PositionUp(userID, in.ID)
	if err != nil {
		return err
	}
	return nil
}
