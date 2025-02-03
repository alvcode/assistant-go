package handler

import (
	"assistant-go/internal/handler"
	dtoUser "assistant-go/internal/layer/dto/user"
	"assistant-go/internal/layer/entity"
	"assistant-go/internal/layer/useCase"
	vmUser "assistant-go/internal/layer/viewModel/user"
	"assistant-go/pkg/vld"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(in entity.User) (*entity.User, error) {
	args := m.Called(in)
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) Find(login string) (*entity.User, error) {
	args := m.Called(login)
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) FindUserToken(token string) (*entity.UserToken, error) {
	args := m.Called(token)
	return args.Get(0).(*entity.UserToken), args.Error(1)
}

func (m *MockUserRepository) SetUserToken(in entity.UserToken) (*entity.UserToken, error) {
	args := m.Called(in)
	return args.Get(0).(*entity.UserToken), args.Error(1)
}

func TestRegisterUserEndpointSuccess(t *testing.T) {
	// Создаем мок-репозиторий
	mockRepo := new(MockUserRepository)
	ctx := context.Background()
	vld.InitValidator(ctx)

	userUseCase := useCase.NewUserUseCase(ctx, mockRepo)

	userHandler := handler.NewUserHandler(userUseCase)

	userEntity := entity.User{ID: 1, Login: "test_user", Password: "test_pwd", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	mockRepo.On("Create", mock.AnythingOfType("entity.User")).Return(&userEntity, nil)
	mockRepo.On("Find", userEntity.Login).Return((*entity.User)(nil), errors.New("user not found"))

	dtoForRequest := dtoUser.LoginAndPassword{Login: "test_user", Password: "test_pwd"}
	// Создаем тестовый HTTP-запрос
	requestBody, _ := json.Marshal(dtoForRequest)
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Эмулируем HTTP-сервер
	rr := httptest.NewRecorder()
	userHandler.Create(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	// Проверяем, что в ответе пришел тот же пользователь
	var responseUser vmUser.User
	err := json.Unmarshal(rr.Body.Bytes(), &responseUser)
	assert.NoError(t, err)

	assert.Equal(t, userEntity.Login, responseUser.Login)
}

func TestRegisterUserEndpointErrorExists(t *testing.T) {
	mockRepo := new(MockUserRepository)
	ctx := context.Background()
	vld.InitValidator(ctx)

	userUseCase := useCase.NewUserUseCase(ctx, mockRepo)

	userHandler := handler.NewUserHandler(userUseCase)

	userEntity := entity.User{ID: 1, Login: "test_user", Password: "test_pwd", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	mockRepo.On("Create", mock.AnythingOfType("entity.User")).Return(&userEntity, nil)
	mockRepo.On("Find", userEntity.Login).Return(&userEntity, nil)

	dtoForRequest := dtoUser.LoginAndPassword{Login: "test_user", Password: "test_pwd"}
	requestBody, _ := json.Marshal(dtoForRequest)
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	userHandler.Create(rr, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
}
