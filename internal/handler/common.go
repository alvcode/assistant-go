package handler

import (
	"assistant-go/internal/config"
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/layer/entity"
	"assistant-go/internal/layer/repository"
	"assistant-go/internal/locale"
	"assistant-go/internal/storage/postgres"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"net"
	"net/http"
	"strings"
	"time"
)

var userRepository repository.UserRepository
var blockIpRepository repository.BlockIPRepository
var blockEventRepository repository.BlockEventRepository
var appConf *config.Config

func InitHandler(ctx context.Context, db *pgxpool.Pool, cfg *config.Config) {
	userRepository = repository.NewUserRepository(ctx, db)
	blockIpRepository = repository.NewBlockIpRepository(ctx, db)
	blockEventRepository = repository.NewBlockEventRepository(ctx, db)
	appConf = cfg
}

var (
	ErrSplitHostIP = errors.New("split host ip fail")
	ErrDetermineIP = errors.New("determine ip fail")
)

type ErrorResponse struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
	Code    int    `json:"code"`
}

func SendErrorResponse(w http.ResponseWriter, message string, status int, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	errorResponse := ErrorResponse{
		Message: message,
		Status:  status,
		Code:    code,
	}

	err := json.NewEncoder(w).Encode(errorResponse)
	if err != nil {
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
	}
}

func SendResponse(w http.ResponseWriter, status int, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		SendErrorResponse(w, "Failed to encode response", http.StatusInternalServerError, 0)
	}
}

func GetAuthUser(r *http.Request) (*entity.User, error) {
	userEntity, ok := r.Context().Value(UserContextKey).(*entity.User)
	if !ok {
		return nil, errors.New("auth user not found")
	}
	return userEntity, nil
}

var (
	BlockEventInputDataType    = "validate_input_data"
	BlockEventDecodeBodyType   = "decode_body"
	BlockEventErrorSignInType  = "sign_in"
	BlockEventUnauthorizedType = "unauthorized"
	BlockEventRefreshTokenType = "refresh_token"
	BlockEventOtherType        = "other"
)

func BlockEventHandle(r *http.Request, eventName string) {
	fmt.Println("block event func")
	IPAddress, err := GetIpAddress(r)
	if err == nil {
		_, err := blockEventRepository.SetEvent(IPAddress, eventName, time.Now().UTC())
		if err != nil {
			return
		}
	}
	fmt.Println(appConf.BlockingParanoia)
}

func GetIpAddress(r *http.Request) (string, error) {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}

	if strings.Contains(IPAddress, ":") {
		host, _, err := net.SplitHostPort(IPAddress)
		if err == nil {
			IPAddress = host
		} else {
			return "", ErrSplitHostIP
		}
	}

	dtoBlockIP := dto.BlockIP{IP: IPAddress}
	if err := dtoBlockIP.Validate("en"); err != nil {
		return "", ErrDetermineIP
	}
	return IPAddress, nil
}

func BuildErrorMessageCommon(lang string, err error) string {
	switch {
	case errors.Is(err, postgres.ErrUnexpectedDBError):
		return locale.T(lang, "unexpected_database_error")
	case errors.Is(err, ErrSplitHostIP):
		return locale.T(lang, "unexpected_error")
	case errors.Is(err, ErrDetermineIP):
		return locale.T(lang, "failed_to_determine_ip")
	default:
		return locale.T(lang, "unexpected_error")
	}
}
