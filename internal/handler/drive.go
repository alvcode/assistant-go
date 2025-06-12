package handler

import (
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/layer/ucase"
	"assistant-go/internal/layer/vmodel"
	"assistant-go/internal/locale"
	"assistant-go/internal/logging"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
)

var ()

type DriveHandler struct {
	useCase ucase.DriveUseCase
}

func NewDriveHandler(useCase ucase.DriveUseCase) *DriveHandler {
	return &DriveHandler{
		useCase: useCase,
	}
}

func (h *DriveHandler) GetTree(w http.ResponseWriter, r *http.Request) {
	langRequest := locale.GetLangFromContext(r.Context())
	authUser, err := GetAuthUser(r)
	if err != nil {
		BlockEventHandle(r, BlockEventUnauthorizedType)
		SendErrorResponse(w, locale.T(langRequest, "unauthorized"), http.StatusUnauthorized, 0)
		return
	}

	var parentID *int
	parentIDStr := r.URL.Query().Get("parentId")

	if parentIDStr != "" {
		parentIDInt, err := strconv.Atoi(parentIDStr)

		if err != nil {
			BlockEventHandle(r, BlockEventInputDataType)
			SendErrorResponse(w, locale.T(langRequest, "parameter_conversion_error"), http.StatusBadRequest, 0)
			return
		}
		parentID = &parentIDInt
	}

	driveStructList, err := h.useCase.GetTree(parentID, authUser)
	if err != nil {
		SendErrorResponse(w, buildErrorMessage(langRequest, err), http.StatusUnprocessableEntity, 0)
		return
	}

	result := vmodel.DriveStructsFromEntities(driveStructList)
	SendResponse(w, http.StatusOK, result)
	return
}

func (h *DriveHandler) CreateDirectory(w http.ResponseWriter, r *http.Request) {
	langRequest := locale.GetLangFromContext(r.Context())
	var createDirectoryDTO dto.DriveCreateDirectory
	authUser, err := GetAuthUser(r)
	if err != nil {
		BlockEventHandle(r, BlockEventUnauthorizedType)
		SendErrorResponse(w, locale.T(langRequest, "unauthorized"), http.StatusUnauthorized, 0)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&createDirectoryDTO)
	if err != nil {
		BlockEventHandle(r, BlockEventDecodeBodyType)
		SendErrorResponse(w, locale.T(langRequest, "error_reading_request_body"), http.StatusBadRequest, 0)
		return
	}

	if err = createDirectoryDTO.Validate(langRequest); err != nil {
		BlockEventHandle(r, BlockEventInputDataType)
		SendErrorResponse(w, fmt.Sprint(err), http.StatusUnprocessableEntity, 0)
		return
	}

	driveStructList, err := h.useCase.CreateDirectory(&createDirectoryDTO, authUser)
	if err != nil {
		SendErrorResponse(w, buildErrorMessage(langRequest, err), http.StatusUnprocessableEntity, 0)
		return
	}

	result := vmodel.DriveStructsFromEntities(driveStructList)
	SendResponse(w, http.StatusCreated, result)
	return
}

func (h *DriveHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
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

	uploadFileDto := dto.DriveUploadFile{
		File:                  file,
		OriginalFilename:      header.Filename,
		MaxSizeBytes:          appConf.Drive.UploadMaxSize << 20,
		StorageMaxSizePerUser: appConf.Drive.LimitPerUser << 20,
		SavePath:              appConf.Drive.SavePath,
	}
}
