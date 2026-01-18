package handler

import (
	"assistant-go/internal/config"
	"assistant-go/internal/handler"
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/layer/entity"
	"assistant-go/internal/layer/repository"
	"assistant-go/internal/layer/ucase"
	"assistant-go/internal/layer/vmodel"
	"assistant-go/internal/logging"
	mocks "assistant-go/mocks/layer/repository"
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

func setupTest() (*mocks.MockUserRepository, *handler.UserHandler, context.Context) {
	mockRepo := &mocks.MockUserRepository{}

	ctx := context.Background()
	logger := logging.NewLogger(logging.TestsEnv)
	ctx = logging.ContextWithLogger(ctx, logger)

	vld.InitValidator(ctx)

	repos := &repository.Repositories{
		UserRepository: mockRepo,
	}

	cfg := &config.Config{BlockingParanoia: 0}
	handler.InitHandler(repos, cfg)

	userUseCase := ucase.NewUserUseCase(repos)
	userHandler := handler.NewUserHandler(userUseCase)

	return mockRepo, userHandler, ctx
}

func TestRegisterUserEndpoint(t *testing.T) {
	mockRepo, userHandler, _ := setupTest()

	nowDatetime := time.Now()
	userEntity := entity.User{ID: 1, Login: "test_user", Password: "test_pwd", CreatedAt: nowDatetime, UpdatedAt: nowDatetime}
	userVM := vmodel.UserFromEnity(&userEntity)
	userVMBytes, _ := json.Marshal(userVM)

	tests := []struct {
		name               string
		requestData        dto.UserLoginAndPassword
		mockSetup          func()
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name: "Success",
			requestData: dto.UserLoginAndPassword{
				Login: "test_user", Password: "test_pwd",
			},
			mockSetup: func() {
				mockRepo.On("Find", mock.Anything, "test_user").Return((*entity.User)(nil), errors.New("user not found"))
				mockRepo.On("Create", mock.Anything, mock.AnythingOfType("entity.User")).Return(&userEntity, nil)
			},
			expectedStatusCode: http.StatusCreated,
			expectedResponse:   string(userVMBytes),
		},
		{
			name: "User already exists",
			requestData: dto.UserLoginAndPassword{
				Login: "test_user", Password: "test_pwd",
			},
			mockSetup: func() {
				mockRepo.On("Find", mock.Anything, "test_user").Return(&userEntity, nil)
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   `{"message":"user_already_exists","status":422,"code":0}`,
		},
		{
			name: "Missing login",
			requestData: dto.UserLoginAndPassword{
				Login: "", Password: "test_pwd",
			},
			mockSetup:          func() {},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   `{"message":"Login is a required field","status":422,"code":0}`,
		},
		{
			name: "Missing Password",
			requestData: dto.UserLoginAndPassword{
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

	expiredToken := int(time.Now().Add(4 * time.Hour).Unix())
	userTokenEntity := entity.UserToken{
		UserId:       1,
		Token:        "token_string_sdjhfgasdjfgasjdgfasjasdgvhajkadhsdnjhsdnjhsfnjsdfg",
		RefreshToken: "refresh_token_ksdjgjasdbghjasgjasghdsfhasdhadhsdfnjhsnjh",
		ExpiredTo:    expiredToken,
	}
	userTokenVM := vmodel.UserTokenFromEnity(&userTokenEntity)
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
		requestData        dto.UserLoginAndPassword
		mockSetup          func()
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name: "Success",
			requestData: dto.UserLoginAndPassword{
				Login: "test_user", Password: "test_pwd",
			},
			mockSetup: func() {
				mockRepo.On("Find", mock.Anything, "test_user").Return(&userEntity, nil)
				mockRepo.On("FindUserToken", mock.Anything, mock.Anything).Return((*entity.UserToken)(nil), pgx.ErrNoRows)
				mockRepo.On("SetUserToken", mock.Anything, mock.AnythingOfType("entity.UserToken")).Return(&userTokenEntity, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   string(userTokenVMBytes),
		},
		{
			name: "Wrong Password",
			requestData: dto.UserLoginAndPassword{
				Login: "test_user", Password: "test_pwdd",
			},
			mockSetup: func() {
				mockRepo.On("Find", mock.Anything, "test_user").Return(&userEntity, nil)
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   `{"message":"incorrect_username_or_password","status":422,"code":0}`,
		},
		{
			name: "Empty Password",
			requestData: dto.UserLoginAndPassword{
				Login: "test_user", Password: "",
			},
			mockSetup: func() {
				mockRepo.On("Find", mock.Anything, "test_user").Return(&userEntity, nil)
				mockRepo.On("FindUserToken", mock.Anything, mock.Anything).Return((*entity.UserToken)(nil), pgx.ErrNoRows)
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   `{"message":"Password is a required field","status":422,"code":0}`,
		},
		{
			name: "Empty Login",
			requestData: dto.UserLoginAndPassword{
				Login: "", Password: "test_pwd",
			},
			mockSetup: func() {
				mockRepo.On("Find", mock.Anything, "test_user").Return(&userEntity, nil)
				mockRepo.On("FindUserToken", mock.Anything, mock.Anything).Return((*entity.UserToken)(nil), pgx.ErrNoRows)
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

	expiredToken := int(time.Now().Add(4 * time.Hour).Unix())
	userTokenEntity := entity.UserToken{
		UserId:       1,
		Token:        "token_string_sdjhfgasdjfgasjdgfasjasdgvhajkadhsdnjhsdnjhsfnjsdfg",
		RefreshToken: "refresh_token_ksdjgjasdbghjasgjasghdsfhasdhadhsdfnjhsnjh",
		ExpiredTo:    expiredToken,
	}
	userTokenVM := vmodel.UserTokenFromEnity(&userTokenEntity)
	userTokenVMBytes, _ := json.Marshal(userTokenVM)

	tests := []struct {
		name               string
		requestData        dto.UserRefreshToken
		mockSetup          func()
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name: "Success",
			requestData: dto.UserRefreshToken{
				Token:        "token_string_sdjhfgasdjfgasjdgfasjasdgvhajkadhsdnjhsdnjhsfnjsdfg",
				RefreshToken: "refresh_token_ksdjgjasdbghjasgjasghdsfhasdhadhsdfnjhsnjh",
			},
			mockSetup: func() {
				mockRepo.On("FindUserToken", mock.Anything, mock.Anything).Return(&userTokenEntity, nil).Once()
				mockRepo.On("FindUserToken", mock.Anything, mock.Anything).Return((*entity.UserToken)(nil), pgx.ErrNoRows).Once()
				mockRepo.On("SetUserToken", mock.Anything, mock.AnythingOfType("entity.UserToken")).Return(&userTokenEntity, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   string(userTokenVMBytes),
		},
		{
			name: "Not Found Token",
			requestData: dto.UserRefreshToken{
				Token:        "token_string_sdjhfgasdjfgasjdgfasjasdgvhajkadhsdnjhsdnjhsfnjsdfg",
				RefreshToken: "refresh_token_ksdjgjasdbghjasgjasghdsfhasdhadhsdfnjhsnjh",
			},
			mockSetup: func() {
				mockRepo.On("FindUserToken", mock.Anything, mock.Anything).Return((*entity.UserToken)(nil), pgx.ErrNoRows).Once()
				mockRepo.On("SetUserToken", mock.Anything, mock.AnythingOfType("entity.UserToken")).Return(&userTokenEntity, nil)
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   `{"message":"refresh_token_not_found","status":422,"code":0}`,
		},
		{
			name: "Wrong Refresh Token",
			requestData: dto.UserRefreshToken{
				Token:        "token_string_sdjhfgasdjfgasjdgfasjasdgvhajkadhsdnjhsdnjhsfnjsdfg",
				RefreshToken: "refresh_token_ksdjgjasdbghjasgjasghdsfhasdhadhsdfnjhsnjhh",
			},
			mockSetup: func() {
				mockRepo.On("FindUserToken", mock.Anything, mock.Anything).Return(&userTokenEntity, nil).Once()
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
