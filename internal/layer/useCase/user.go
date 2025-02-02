package useCase

import (
	"assistant-go/internal/layer/dto/user"
	"assistant-go/internal/layer/entity"
	"assistant-go/internal/layer/repository"
	"assistant-go/internal/locale"
	"assistant-go/internal/logging"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type UserUseCase interface {
	Create(in dtoUser.LoginAndPassword, lang string) (*entity.User, error)
	Login(in dtoUser.LoginAndPassword, lang string) (*entity.UserToken, error)
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

func (uc *userUseCase) Create(in dtoUser.LoginAndPassword, lang string) (*entity.User, error) {
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

func (uc *userUseCase) Login(in dtoUser.LoginAndPassword, lang string) (*entity.UserToken, error) {
	existingUser, _ := uc.userRepository.Find(in.Login)
	if existingUser == nil {
		return nil, errors.New(locale.T(lang, "incorrect_username_or_password"))
	}

	err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(in.Password))
	if err != nil {
		return nil, errors.New(locale.T(lang, "incorrect_username_or_password"))
	}

	var userTokenEntity entity.UserToken
	for {
		token := GenerateAPIToken()
		refreshToken := GenerateAPIToken()

		existingToken, _ := uc.userRepository.FindUserToken(token)
		if existingToken == nil {
			userTokenEntity = entity.UserToken{
				UserId:       existingUser.ID,
				Token:        token,
				RefreshToken: refreshToken,
				ExpiredTo:    uint32(time.Now().Add(4 * time.Hour).Unix()),
			}
			break
		}
	}
	data, err := uc.userRepository.SetUserToken(userTokenEntity)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, errors.New(locale.T(lang, "unexpected_database_error"))
	}
	return data, nil
}

func GenerateAPIToken() string {
	bytes := make([]byte, 48)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(bytes)
}
