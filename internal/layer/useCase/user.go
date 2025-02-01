package useCase

import (
	dtoUser "assistant-go/internal/layer/dto/user"
	"assistant-go/internal/layer/entity"
	"assistant-go/internal/layer/repository"
	"assistant-go/internal/locale"
	"assistant-go/internal/logging"
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type UserUseCase interface {
	Create(in dtoUser.CreateDto, lang string) (*entity.User, error)
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

func (uc *userUseCase) Create(in dtoUser.CreateDto, lang string) (*entity.User, error) {
	//err := bcrypt.CompareHashAndPassword([]byte(storedHashedPassword), []byte(inputPassword))
	existingUser, _ := uc.userRepository.Find(in.Login)
	if existingUser != nil {
		return nil, errors.New(locale.T(lang, "user_already_exists"))
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(in.Password), 11)
	if err != nil {
		return nil, err
	}

	userEntity := entity.User{
		Login:     in.Login,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	data, err := uc.userRepository.Create(userEntity)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, errors.New(locale.T(lang, "unexpected_database_error"))
	}
	return data, nil
}
