package useCase

import (
	"assistant-go/internal/layer/entity"
	"assistant-go/internal/layer/repository"
)

type UserUseCase interface {
	Create(in entity.User) (*entity.User, error)
}

type userUseCase struct {
	userRepository repository.UserRepository
}

func NewUserUseCase(userRepository repository.UserRepository) UserUseCase {
	return &userUseCase{
		userRepository: userRepository,
	}
}

func (uc *userUseCase) Create(in entity.User) (*entity.User, error) {
	data, err := uc.userRepository.Create(in)
	if err != nil {
		return nil, err
	}
	return data, nil
}
