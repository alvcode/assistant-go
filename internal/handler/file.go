package handler

import (
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/layer/repository"
	"assistant-go/internal/layer/ucase"
	"assistant-go/internal/layer/vmodel"
	"assistant-go/internal/locale"
	"assistant-go/internal/logging"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
)

var (
	ErrFileInvalidReadForm = errors.New("invalid read form")
)

type FileHandler struct {
	useCase ucase.FileUseCase
}

func NewFileHandler(useCase ucase.FileUseCase) *FileHandler {
	return &FileHandler{
		useCase: useCase,
	}
}

func (h *FileHandler) Upload(w http.ResponseWriter, r *http.Request) {
	langRequest := locale.GetLangFromContext(r.Context())

	authUser, err := GetAuthUser(r)
	if err != nil {
		BlockEventHandle(r, BlockEventUnauthorizedType)
		SendErrorResponse(w, locale.T(langRequest, "unauthorized"), http.StatusUnauthorized, 0)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		logging.GetLogger(r.Context()).Error(err)
		BlockEventHandle(r, BlockEventDecodeBodyType)
		SendErrorResponse(w, buildErrorMessage(langRequest, ErrFileInvalidReadForm), http.StatusUnprocessableEntity, 0)
		return
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			SendErrorResponse(w, "failed to close uploaded file", http.StatusUnprocessableEntity, 0)
			return
		}
	}(file)

	uploadFileDto := dto.UploadFile{
		File:             file,
		OriginalFilename: header.Filename,
		MaxSizeBytes:     appConf.File.UploadMaxSize << 20,
		StorageMaxSize:   appConf.File.LimitStoragePerUser << 20,
		SavePath:         appConf.File.SavePath,
	}

	upload, err := h.useCase.Upload(uploadFileDto, authUser)
	if err != nil {
		switch {
		case errors.Is(err, ucase.ErrFileTooLarge),
			errors.Is(err, ucase.ErrFileInvalidType),
			errors.Is(err, ucase.ErrFileExtensionDoesNotMatch),
			errors.Is(err, ucase.ErrFileNotSafeFilename):
			BlockEventHandle(r, BlockEventInputDataType)
		default:
			BlockEventHandle(r, BlockEventOtherType)
		}

		var errorMsg string
		if errors.Is(err, ucase.ErrFileInvalidType) || errors.Is(err, ucase.ErrFileExtensionDoesNotMatch) {
			errorMsg = locale.T(
				langRequest,
				"file_invalid_type",
				map[string]interface{}{"Formats": strings.Join(h.useCase.GetAllowedExtensions(), ", ")},
			)
		} else {
			errorMsg = buildErrorMessage(langRequest, err)
		}
		SendErrorResponse(w, errorMsg, http.StatusUnprocessableEntity, 0)
		return
	}

	uploadUrl := appConf.ThisServiceDomain + "/api/files/hash/" + upload.Hash

	result := vmodel.FileFromEntity(upload, uploadUrl)
	SendResponse(w, http.StatusCreated, result)
	return
}

func (h *FileHandler) GetByHash(w http.ResponseWriter, r *http.Request) {
	langRequest := locale.GetLangFromContext(r.Context())
	var fileHashDto dto.GetFileByHash

	params := httprouter.ParamsFromContext(r.Context())
	fileHashDto.Hash = params.ByName("hash")

	if err := fileHashDto.Validate(langRequest); err != nil {
		BlockEventHandle(r, BlockEventInputDataType)
		SendErrorResponse(w, fmt.Sprint(err), http.StatusUnprocessableEntity, 0)
		return
	}

	fileHashDto.SavePath = appConf.File.SavePath
	fileDto, err := h.useCase.GetFileByHash(fileHashDto)
	if err != nil {
		var responseStatus int
		if errors.Is(err, ucase.ErrFileNotFound) {
			responseStatus = http.StatusNotFound
			BlockEventHandle(r, BlockEventFileNotFoundType)
		} else if errors.Is(err, repository.ErrFileNotFoundInFilesystem) {
			responseStatus = http.StatusNotFound
		} else {
			responseStatus = http.StatusUnprocessableEntity
		}
		SendErrorResponse(w, buildErrorMessage(langRequest, err), responseStatus, 0)
		return
	}
	defer func() {
		if closer, ok := fileDto.File.(io.Closer); ok {
			_ = closer.Close()
		}
	}()

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileDto.OriginalFilename))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)

	_, err = io.Copy(w, fileDto.File)
	if err != nil {
		SendErrorResponse(
			w,
			fmt.Sprintf("%s: %v", locale.T(langRequest, "file_failed_to_send"), err),
			http.StatusInternalServerError,
			0,
		)
		return
	}
	return
}
