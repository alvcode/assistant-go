package service

import (
	"assistant-go/internal/layer/entity"
	"assistant-go/internal/layer/repository"
	"assistant-go/internal/logging"
	"assistant-go/internal/storage/postgres"
	"context"
	"errors"
	"sort"
)

var (
	ErrCategoryNotFound             = errors.New("category not found")
	ErrCategoryAlreadyFirstPosition = errors.New("category already first position")
)

type PositionService interface {
	CalculateForNew(userID int, parentCatID *int) (int, error)
	PositionUp(userID int, catID int) error
}

type positionService struct {
	ctx          context.Context
	repositories *repository.Repositories
}

func (ps *positionService) CalculateForNew(userID int, parentCatID *int) (int, error) {
	maxPosition, err := ps.repositories.NoteCategoryRepository.GetMaxPosition(ps.ctx, userID, parentCatID)
	if err != nil {
		return 0, err
	}
	return maxPosition + 1, nil
}

func (ps *positionService) PositionUp(userID int, catID int) error {
	categories, err := ps.repositories.NoteCategoryRepository.FindAll(ps.ctx, userID)
	if err != nil {
		logging.GetLogger(ps.ctx).Error(err)
		return postgres.ErrUnexpectedDBError
	}

	if len(categories) == 0 {
		return ErrCategoryNotFound
	}

	sort.Slice(categories, func(i, j int) bool {
		return categories[i].ID < categories[j].ID
	})

	grouped := make(map[int][]*entity.NoteCategory)
	var needUpCategoryKey int

	// Группируем категории по ParentID
	for _, category := range categories {
		key := 0
		if category.ParentId != nil {
			key = *category.ParentId
		}
		grouped[key] = append(grouped[key], category)

		if category.ID == catID {
			needUpCategoryKey = key
		}
	}

	// Сортируем группы по позиции
	for key := range grouped {
		sort.Slice(grouped[key], func(i, j int) bool {
			return grouped[key][i].Position < grouped[key][j].Position
		})
	}

	if grouped[needUpCategoryKey][0].ID == catID {
		return ErrCategoryAlreadyFirstPosition
	}

	// Поиск и обмен местами
	for i, category := range grouped[needUpCategoryKey] {
		if category.ID == catID {
			grouped[needUpCategoryKey][i], grouped[needUpCategoryKey][i-1] =
				grouped[needUpCategoryKey][i-1], grouped[needUpCategoryKey][i]
			break
		}
	}

	// Пересчет позиций
	for key := range grouped {
		for i := range grouped[key] {
			newPosition := i + 1
			if grouped[key][i].Position != newPosition {
				grouped[key][i].Position = newPosition
				err = ps.repositories.NoteCategoryRepository.UpdatePosition(ps.ctx, grouped[key][i])
				if err != nil {
					logging.GetLogger(ps.ctx).Error(err)
					return postgres.ErrUnexpectedDBError
				}
			}
		}
	}

	return nil
}
