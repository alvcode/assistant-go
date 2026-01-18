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
	BlockEventFileNotFoundType = "file_not_found"
	BlockEventOtherType        = "other"
	BlockEventTooManyRequests  = "too_many_requests"
)

func BlockEventHandle(r *http.Request, eventName string) {
	if appConf.BlockingParanoia == 0 {
		return
	}

	IPAddress, err := GetIpAddress(r)
	if err == nil {
		_, err := blockEventRepository.SetEvent(r.Context(), IPAddress, eventName, time.Now().UTC())
		if err != nil {
			return
		}
	}
	checkTime := time.Now().Add(-30 * time.Minute).UTC()

	blockEventStat, err := blockEventRepository.GetStat(r.Context(), IPAddress, checkTime)
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
	var fileNotFoundMaxCount int
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
		fileNotFoundMaxCount = 40
		tooManyRequestsMaxCount = 60
	case 2:
		blockMinute = 420 // 7 hour
		allMaxCount = 150
		validateInputMaxCount = 40
		decodeBodyMaxCount = 20
		signInMaxCount = 20
		unauthorizedMaxCount = 30
		refreshTokenMaxCount = 50
		pageNotFoundMaxCount = 10
		fileNotFoundMaxCount = 20
		tooManyRequestsMaxCount = 30
	case 3:
		blockMinute = 2880 // 2 day
		allMaxCount = 70
		validateInputMaxCount = 30
		decodeBodyMaxCount = 10
		signInMaxCount = 10
		unauthorizedMaxCount = 20
		refreshTokenMaxCount = 30
		pageNotFoundMaxCount = 5
		fileNotFoundMaxCount = 10
		tooManyRequestsMaxCount = 15
	}

	if blockEventStat.All >= allMaxCount ||
		blockEventStat.ValidateInputData >= validateInputMaxCount ||
		blockEventStat.DecodeBody >= decodeBodyMaxCount ||
		blockEventStat.SignIn >= signInMaxCount ||
		blockEventStat.Unauthorized >= unauthorizedMaxCount ||
		blockEventStat.RefreshToken >= refreshTokenMaxCount ||
		blockEventStat.PageNotFound >= pageNotFoundMaxCount ||
		blockEventStat.FileNotFound >= fileNotFoundMaxCount ||
		blockEventStat.TooManyRequests >= tooManyRequestsMaxCount {
		unblockTime := time.Now().Add(time.Duration(blockMinute) * time.Minute).UTC()
		_ = blockIpRepository.SetBlock(r.Context(), IPAddress, unblockTime)
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
	case errors.Is(err, ucase.ErrFileTooLarge):
		return locale.T(lang, "file_too_large")
	case errors.Is(err, ucase.ErrFileReading):
		return locale.T(lang, "file_error_reading")
	case errors.Is(err, ucase.ErrFileInvalidType):
		return locale.T(lang, "file_invalid_type")
	case errors.Is(err, ucase.ErrFileResettingPointer):
		return locale.T(lang, "file_error_resetting_pointer")
	case errors.Is(err, ucase.ErrFileUnableToSeek):
		return locale.T(lang, "file_unable_to_seek")
	case errors.Is(err, ucase.ErrFileExtensionDoesNotMatch):
		return locale.T(lang, "file_extension_does_not_match")
	case errors.Is(err, ucase.ErrFileNotSafeFilename):
		return locale.T(lang, "file_not_safe_filename")
	case errors.Is(err, ucase.ErrFileSave):
		return locale.T(lang, "file_error_save")
	case errors.Is(err, repository.ErrFileSave):
		return locale.T(lang, "file_error_save")
	case errors.Is(err, ErrFileInvalidReadForm):
		return locale.T(lang, "file_error_reading")
	case errors.Is(err, ucase.ErrFileNotFound):
		return locale.T(lang, "file_not_found")
	case errors.Is(err, repository.ErrFileNotFoundInFilesystem):
		return locale.T(lang, "file_not_found_in_filesystem")
	case errors.Is(err, ucase.ErrFileSystemIsFull):
		return locale.T(lang, "file_system_is_full")
	case errors.Is(err, ucase.ErrDriveDirectoryExists):
		return locale.T(lang, "drive_dir_exists")
	case errors.Is(err, ucase.ErrDriveParentIdNotFound):
		return locale.T(lang, "drive_parent_id_not_found")
	case errors.Is(err, ucase.ErrDriveFileTooLarge):
		return locale.T(lang, "file_too_large")
	case errors.Is(err, ucase.ErrDriveFileTooLargeUseChunks):
		return locale.T(lang, "file_too_large_use_chunks")
	case errors.Is(err, ucase.ErrDriveUnavailableForChunks):
		return locale.T(lang, "drive_method_unavailable_for_chunks")
	case errors.Is(err, ucase.ErrDriveFileSystemIsFull):
		return locale.T(lang, "file_system_is_full")
	case errors.Is(err, ucase.ErrDriveFileNotSafeFilename):
		return locale.T(lang, "file_not_safe_filename")
	case errors.Is(err, ucase.ErrDriveFileSave):
		return locale.T(lang, "file_error_save")
	case errors.Is(err, ucase.ErrDriveStructNotFound):
		return locale.T(lang, "drive_struct_not_found")
	case errors.Is(err, ucase.ErrDriveFilenameExists):
		return locale.T(lang, "drive_filename_exists")
	case errors.Is(err, ucase.ErrDriveRelocatableStructureNotFound):
		return locale.T(lang, "drive_relocatable_structure_not_found")
	case errors.Is(err, ucase.ErrDriveMovingIntoOneself):
		return locale.T(lang, "drive_moving_into_oneself")
	case errors.Is(err, ucase.ErrDriveParentRefOfTheRelocatableStruct):
		return locale.T(lang, "drive_parent_references_one_of_the_relocatable_struct")
	case errors.Is(err, ucase.ErrDriveEncrypting):
		return locale.T(lang, "drive_encryption_error")
	case errors.Is(err, ucase.ErrDriveDecrypting):
		return locale.T(lang, "drive_decryption_error")
	case errors.Is(err, ucase.ErrNoteShareExists):
		return locale.T(lang, "note_share_exists")
	default:
		return locale.T(lang, "unexpected_error")
	}
}
