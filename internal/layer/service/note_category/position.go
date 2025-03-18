package service

import (
	"assistant-go/internal/layer/entity"
	"assistant-go/internal/layer/repository"
	"assistant-go/internal/locale"
	"assistant-go/internal/logging"
	"context"
	"errors"
	"sort"
)

type PositionService interface {
	CalculateForNew(userID int, parentCatID *int) (int, error)
	PositionUp(userID int, catID int, lang string) error
}

type positionService struct {
	ctx          context.Context
	repositories *repository.Repositories
}

func (ps *positionService) CalculateForNew(userID int, parentCatID *int) (int, error) {
	maxPosition, err := ps.repositories.NoteCategoryRepository.GetMaxPosition(userID, parentCatID)
	if err != nil {
		return 0, err
	}
	return maxPosition + 1, nil
}

func (ps *positionService) PositionUp(userID int, catID int, lang string) error {
	categories, err := ps.repositories.NoteCategoryRepository.FindAll(userID)
	if err != nil {
		logging.GetLogger(ps.ctx).Error(err)
		return errors.New(locale.T(lang, "unexpected_database_error"))
	}

	if len(categories) == 0 {
		return errors.New(locale.T(lang, "category_not_found"))
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
		return errors.New(locale.T(lang, "category_already_in_1_position"))
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
				err = ps.repositories.NoteCategoryRepository.UpdatePosition(grouped[key][i])
				if err != nil {
					logging.GetLogger(ps.ctx).Error(err)
					return errors.New(locale.T(lang, "unexpected_database_error"))
				}
			}
		}
	}

	return nil
}
