package service

import (
	"assistant-go/internal/layer/repository"
	"context"
)

type PositionService interface {
	CalculateForNew(userID int, parentCatID *int) int
}

type positionService struct {
	ctx          context.Context
	repositories *repository.Repositories
}

func (ps *positionService) CalculateForNew(userID int, parentCatID *int) int {
	// если parentCatID есть, то получаем max position среди тех у кого такая же родительская категория
	// если parentCatID нет, то получаем max position БЕЗ родительской
	// делаем +1, возвращаем
}
