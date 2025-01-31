package useCase

import (
	dtoUser "assistant-go/internal/layer/dto/user"
	"assistant-go/internal/layer/entity"
	"assistant-go/internal/layer/repository"
	"context"
)

type UserUseCase interface {
	Create(in dtoUser.CreateDto) (*entity.User, error)
}

type userUseCase struct {
	ctx            context.Context
	userRepository repository.UserRepository
}

func NewUserUseCase(ctx context.Context, userRepository repository.UserRepository) UserUseCase {
	return &userUseCase{
		ctx:            ctx,
		userRepository: userRepository,
	}
}

func (uc *userUseCase) Create(in dtoUser.CreateDto) (*entity.User, error) {
	userEntity := entity.User{
		Login:    in.Login,
		Password: in.Password,
	}
	data, err := uc.userRepository.Create(userEntity)
	if err != nil {
		return nil, err
	}
	return data, nil
}
