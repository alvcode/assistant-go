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
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
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

func (m *MockUserRepository) FindById(id int) (*entity.User, error) {
	args := m.Called(id)
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

func setupTest() (*MockUserRepository, *handler.UserHandler, context.Context) {
	mockRepo := new(MockUserRepository)
	ctx := context.Background()
	vld.InitValidator(ctx)

	userUseCase := useCase.NewUserUseCase(ctx, mockRepo)
	userHandler := handler.NewUserHandler(userUseCase)

	return mockRepo, userHandler, ctx
}

func TestRegisterUserEndpoint(t *testing.T) {
	mockRepo, userHandler, _ := setupTest()

	nowDatetime := time.Now()
	userEntity := entity.User{ID: 1, Login: "test_user", Password: "test_pwd", CreatedAt: nowDatetime, UpdatedAt: nowDatetime}
	userVM := vmUser.UserVMFromEnity(&userEntity)
	userVMBytes, _ := json.Marshal(userVM)

	tests := []struct {
		name               string
		requestData        dtoUser.LoginAndPassword
		mockSetup          func()
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name: "Success",
			requestData: dtoUser.LoginAndPassword{
				Login: "test_user", Password: "test_pwd",
			},
			mockSetup: func() {
				mockRepo.On("Find", "test_user").Return((*entity.User)(nil), errors.New("user not found"))
				mockRepo.On("Create", mock.AnythingOfType("entity.User")).Return(&userEntity, nil)
			},
			expectedStatusCode: http.StatusCreated,
			expectedResponse:   string(userVMBytes),
		},
		{
			name: "User already exists",
			requestData: dtoUser.LoginAndPassword{
				Login: "test_user", Password: "test_pwd",
			},
			mockSetup: func() {
				mockRepo.On("Find", "test_user").Return(&userEntity, nil)
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   `{"message":"user_already_exists","status":422,"code":0}`,
		},
		{
			name: "Missing login",
			requestData: dtoUser.LoginAndPassword{
				Login: "", Password: "test_pwd",
			},
			mockSetup:          func() {},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   `{"message":"Login is a required field","status":422,"code":0}`,
		},
		{
			name: "Missing Password",
			requestData: dtoUser.LoginAndPassword{
				Login: "test_user", Password: "",
			},
			mockSetup:          func() {},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   `{"message":"Password is a required field","status":422,"code":0}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil // Очищаем вызовы перед тестом
			tt.mockSetup()

			requestBody, _ := json.Marshal(tt.requestData)
			req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			userHandler.Create(rr, req)

			assert.Equal(t, tt.expectedStatusCode, rr.Code)
			assert.JSONEq(t, tt.expectedResponse, rr.Body.String())
		})
	}
}

func TestLoginEndpoint(t *testing.T) {
	mockRepo, userHandler, _ := setupTest()

	expiredToken := uint32(time.Now().Add(4 * time.Hour).Unix())
	userTokenEntity := entity.UserToken{
		UserId:       1,
		Token:        "token_string_sdjhfgasdjfgasjdgfasjasdgvhajkadhsdnjhsdnjhsfnjsdfg",
		RefreshToken: "refresh_token_ksdjgjasdbghjasgjasghdsfhasdhadhsdfnjhsnjh",
		ExpiredTo:    expiredToken,
	}
	userTokenVM := vmUser.UserTokenVMFromEnity(&userTokenEntity)
	userTokenVMBytes, _ := json.Marshal(userTokenVM)

	testPasswordHash, _ := bcrypt.GenerateFromPassword([]byte("test_pwd"), 11)

	nowDatetime := time.Now()
	userEntity := entity.User{
		ID:        1,
		Login:     "test_user",
		Password:  string(testPasswordHash),
		CreatedAt: nowDatetime,
		UpdatedAt: nowDatetime,
	}

	tests := []struct {
		name               string
		requestData        dtoUser.LoginAndPassword
		mockSetup          func()
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name: "Success",
			requestData: dtoUser.LoginAndPassword{
				Login: "test_user", Password: "test_pwd",
			},
			mockSetup: func() {
				mockRepo.On("Find", "test_user").Return(&userEntity, nil)
				mockRepo.On("FindUserToken", mock.Anything).Return((*entity.UserToken)(nil), pgx.ErrNoRows)
				mockRepo.On("SetUserToken", mock.AnythingOfType("entity.UserToken")).Return(&userTokenEntity, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   string(userTokenVMBytes),
		},
		{
			name: "Wrong Password",
			requestData: dtoUser.LoginAndPassword{
				Login: "test_user", Password: "test_pwdd",
			},
			mockSetup: func() {
				mockRepo.On("Find", "test_user").Return(&userEntity, nil)
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   `{"message":"incorrect_username_or_password","status":422,"code":0}`,
		},
		{
			name: "Empty Password",
			requestData: dtoUser.LoginAndPassword{
				Login: "test_user", Password: "",
			},
			mockSetup: func() {
				mockRepo.On("Find", "test_user").Return(&userEntity, nil)
				mockRepo.On("FindUserToken", mock.Anything).Return((*entity.UserToken)(nil), pgx.ErrNoRows)
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   `{"message":"Password is a required field","status":422,"code":0}`,
		},
		{
			name: "Empty Login",
			requestData: dtoUser.LoginAndPassword{
				Login: "", Password: "test_pwd",
			},
			mockSetup: func() {
				mockRepo.On("Find", "test_user").Return(&userEntity, nil)
				mockRepo.On("FindUserToken", mock.Anything).Return((*entity.UserToken)(nil), pgx.ErrNoRows)
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   `{"message":"Login is a required field","status":422,"code":0}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			tt.mockSetup()

			requestBody, _ := json.Marshal(tt.requestData)
			req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			userHandler.Login(rr, req)

			assert.Equal(t, tt.expectedStatusCode, rr.Code)
			assert.JSONEq(t, tt.expectedResponse, rr.Body.String())
		})
	}
}

func TestRefreshTokenEndpoint(t *testing.T) {
	mockRepo, userHandler, _ := setupTest()

	expiredToken := uint32(time.Now().Add(4 * time.Hour).Unix())
	userTokenEntity := entity.UserToken{
		UserId:       1,
		Token:        "token_string_sdjhfgasdjfgasjdgfasjasdgvhajkadhsdnjhsdnjhsfnjsdfg",
		RefreshToken: "refresh_token_ksdjgjasdbghjasgjasghdsfhasdhadhsdfnjhsnjh",
		ExpiredTo:    expiredToken,
	}
	userTokenVM := vmUser.UserTokenVMFromEnity(&userTokenEntity)
	userTokenVMBytes, _ := json.Marshal(userTokenVM)

	tests := []struct {
		name               string
		requestData        dtoUser.RefreshToken
		mockSetup          func()
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name: "Success",
			requestData: dtoUser.RefreshToken{
				Token:        "token_string_sdjhfgasdjfgasjdgfasjasdgvhajkadhsdnjhsdnjhsfnjsdfg",
				RefreshToken: "refresh_token_ksdjgjasdbghjasgjasghdsfhasdhadhsdfnjhsnjh",
			},
			mockSetup: func() {
				mockRepo.On("FindUserToken", mock.Anything).Return(&userTokenEntity, nil).Once()
				mockRepo.On("FindUserToken", mock.Anything).Return((*entity.UserToken)(nil), pgx.ErrNoRows).Once()
				mockRepo.On("SetUserToken", mock.AnythingOfType("entity.UserToken")).Return(&userTokenEntity, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   string(userTokenVMBytes),
		},
		{
			name: "Not Found Token",
			requestData: dtoUser.RefreshToken{
				Token:        "token_string_sdjhfgasdjfgasjdgfasjasdgvhajkadhsdnjhsdnjhsfnjsdfg",
				RefreshToken: "refresh_token_ksdjgjasdbghjasgjasghdsfhasdhadhsdfnjhsnjh",
			},
			mockSetup: func() {
				mockRepo.On("FindUserToken", mock.Anything).Return((*entity.UserToken)(nil), pgx.ErrNoRows).Once()
				mockRepo.On("SetUserToken", mock.AnythingOfType("entity.UserToken")).Return(&userTokenEntity, nil)
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   `{"message":"refresh_token_not_found","status":422,"code":0}`,
		},
		{
			name: "Wrong Refresh Token",
			requestData: dtoUser.RefreshToken{
				Token:        "token_string_sdjhfgasdjfgasjdgfasjasdgvhajkadhsdnjhsdnjhsfnjsdfg",
				RefreshToken: "refresh_token_ksdjgjasdbghjasgjasghdsfhasdhadhsdfnjhsnjhh",
			},
			mockSetup: func() {
				mockRepo.On("FindUserToken", mock.Anything).Return(&userTokenEntity, nil).Once()
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   `{"message":"refresh_token_not_found","status":422,"code":0}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			tt.mockSetup()

			requestBody, _ := json.Marshal(tt.requestData)
			req := httptest.NewRequest("POST", "/api/auth/refresh-token", bytes.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			userHandler.RefreshToken(rr, req)

			assert.Equal(t, tt.expectedStatusCode, rr.Code)
			assert.JSONEq(t, tt.expectedResponse, rr.Body.String())
		})
	}
}
