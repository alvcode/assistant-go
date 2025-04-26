package handler

import (
	"assistant-go/internal/config"
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/layer/entity"
	"assistant-go/internal/layer/repository"
	service "assistant-go/internal/layer/service/note_category"
	"assistant-go/internal/layer/ucase"
	"assistant-go/internal/locale"
	"assistant-go/internal/storage/postgres"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"
)

var userRepository repository.UserRepository
var blockIpRepository repository.BlockIPRepository
var blockEventRepository repository.BlockEventRepository
var rateLimiterRepository repository.RateLimiterRepository
var appConf *config.Config

func InitHandler(repos *repository.Repositories, cfg *config.Config) {
	userRepository = repos.UserRepository
	blockIpRepository = repos.BlockIPRepository
	blockEventRepository = repos.BlockEventRepository
	rateLimiterRepository = repos.RateLimiterRepository
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

func PageNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	BlockEventHandle(r, BlockEventPageNotFoundType)
	SendErrorResponse(w, "Not Found", http.StatusNotFound, 0)
	return
}

var (
	BlockEventInputDataType    = "validate_input_data"
	BlockEventDecodeBodyType   = "decode_body"
	BlockEventErrorSignInType  = "sign_in"
	BlockEventUnauthorizedType = "unauthorized"
	BlockEventRefreshTokenType = "refresh_token"
	BlockEventPageNotFoundType = "page_not_found"
	BlockEventOtherType        = "other"
	BlockEventTooManyRequests  = "too_many_requests"
)

func BlockEventHandle(r *http.Request, eventName string) {
	if appConf.BlockingParanoia == 0 {
		return
	}

	IPAddress, err := GetIpAddress(r)
	if err == nil {
		_, err := blockEventRepository.SetEvent(IPAddress, eventName, time.Now().UTC())
		if err != nil {
			return
		}
	}
	checkTime := time.Now().Add(-30 * time.Minute).UTC()

	blockEventStat, err := blockEventRepository.GetStat(IPAddress, checkTime)
	if err != nil {
		return
	}
	var blockMinute int
	var allMaxCount int
	var validateInputMaxCount int
	var decodeBodyMaxCount int
	var signInMaxCount int
	var unauthorizedMaxCount int
	var refreshTokenMaxCount int
	var pageNotFoundMaxCount int
	var tooManyRequestsMaxCount int
	switch appConf.BlockingParanoia {
	case 1:
		blockMinute = 30
		allMaxCount = 300
		validateInputMaxCount = 60
		decodeBodyMaxCount = 30
		signInMaxCount = 30
		unauthorizedMaxCount = 50
		refreshTokenMaxCount = 70
		pageNotFoundMaxCount = 50
		tooManyRequestsMaxCount = 30
	case 2:
		blockMinute = 420 // 7 hour
		allMaxCount = 150
		validateInputMaxCount = 40
		decodeBodyMaxCount = 20
		signInMaxCount = 20
		unauthorizedMaxCount = 30
		refreshTokenMaxCount = 50
		pageNotFoundMaxCount = 10
		tooManyRequestsMaxCount = 7
	case 3:
		blockMinute = 2880 // 2 day
		allMaxCount = 70
		validateInputMaxCount = 30
		decodeBodyMaxCount = 10
		signInMaxCount = 10
		unauthorizedMaxCount = 20
		refreshTokenMaxCount = 30
		pageNotFoundMaxCount = 5
		tooManyRequestsMaxCount = 3
	}

	if blockEventStat.All >= allMaxCount ||
		blockEventStat.ValidateInputData >= validateInputMaxCount ||
		blockEventStat.DecodeBody >= decodeBodyMaxCount ||
		blockEventStat.SignIn >= signInMaxCount ||
		blockEventStat.Unauthorized >= unauthorizedMaxCount ||
		blockEventStat.RefreshToken >= refreshTokenMaxCount ||
		blockEventStat.TooManyRequests <= tooManyRequestsMaxCount ||
		blockEventStat.PageNotFound >= pageNotFoundMaxCount {
		unblockTime := time.Now().Add(time.Duration(blockMinute) * time.Minute).UTC()
		_ = blockIpRepository.SetBlock(IPAddress, unblockTime)

	}
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

func buildErrorMessage(lang string, err error) string {
	switch {
	case errors.Is(err, postgres.ErrUnexpectedDBError):
		return locale.T(lang, "unexpected_database_error")
	case errors.Is(err, ErrSplitHostIP):
		return locale.T(lang, "unexpected_error")
	case errors.Is(err, ucase.ErrUnexpectedError):
		return locale.T(lang, "unexpected_error")
	case errors.Is(err, ErrDetermineIP):
		return locale.T(lang, "failed_to_determine_ip")
	case errors.Is(err, ucase.ErrUserIncorrectUsernameOrPassword):
		return locale.T(lang, "incorrect_username_or_password")
	case errors.Is(err, ucase.ErrUserAlreadyExists):
		return locale.T(lang, "user_already_exists")
	case errors.Is(err, ucase.ErrRefreshTokenNotFound):
		return locale.T(lang, "refresh_token_not_found")
	case errors.Is(err, ucase.ErrUserNotFound):
		return locale.T(lang, "user_not_found")
	case errors.Is(err, ucase.ErrUserPasswordsAreNotIdentical):
		return locale.T(lang, "passwords_are_not_identical")
	case errors.Is(err, ucase.ErrCategoryParentIdNotFound):
		return locale.T(lang, "parent_id_of_the_category_not_found")
	case errors.Is(err, ucase.ErrCategoryNotFound):
		return locale.T(lang, "category_not_found")
	case errors.Is(err, service.ErrCategoryNotFound):
		return locale.T(lang, "category_not_found")
	case errors.Is(err, ucase.ErrCategoryHasNotes):
		return locale.T(lang, "category_has_notes")
	case errors.Is(err, service.ErrCategoryAlreadyFirstPosition):
		return locale.T(lang, "category_already_in_1_position")
	case errors.Is(err, ucase.ErrNoteNotFound):
		return locale.T(lang, "note_not_found")
	default:
		return locale.T(lang, "unexpected_error")
	}
}
