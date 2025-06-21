package handler

import (
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/layer/ucase"
	"assistant-go/internal/layer/vmodel"
	"assistant-go/internal/locale"
	"assistant-go/internal/logging"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
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
		ParentID:              parentID,
	}

	driveStructList, err := h.useCase.UploadFile(uploadFileDto, authUser)
	if err != nil {
		switch {
		case errors.Is(err, ucase.ErrDriveParentIdNotFound),
			errors.Is(err, ucase.ErrDriveFileNotSafeFilename),
			errors.Is(err, ucase.ErrDriveFileTooLarge):
			BlockEventHandle(r, BlockEventInputDataType)
		default:
			BlockEventHandle(r, BlockEventOtherType)
		}

		SendErrorResponse(w, buildErrorMessage(langRequest, err), http.StatusUnprocessableEntity, 0)
		return
	}

	result := vmodel.DriveStructsFromEntities(driveStructList)
	SendResponse(w, http.StatusCreated, result)
	return
}

func (h *DriveHandler) Delete(w http.ResponseWriter, r *http.Request) {
	langRequest := locale.GetLangFromContext(r.Context())

	authUser, err := GetAuthUser(r)
	if err != nil {
		BlockEventHandle(r, BlockEventUnauthorizedType)
		SendErrorResponse(w, locale.T(langRequest, "unauthorized"), http.StatusUnauthorized, 0)
		return
	}

	var structID int

	params := httprouter.ParamsFromContext(r.Context())
	if structIDStr := params.ByName("id"); structIDStr != "" {
		noteIDInt, err := strconv.Atoi(structIDStr)

		if err != nil {
			BlockEventHandle(r, BlockEventInputDataType)
			SendErrorResponse(w, locale.T(langRequest, "parameter_conversion_error"), http.StatusBadRequest, 0)
			return
		}
		structID = noteIDInt
	}

	err = h.useCase.Delete(structID, appConf.Drive.SavePath, authUser)
	if err != nil {
		//BlockEventHandle(r, BlockEventUnauthorizedType)
		SendErrorResponse(w, buildErrorMessage(langRequest, err), http.StatusUnprocessableEntity, 0)
		return
	}

	SendResponse(w, http.StatusNoContent, nil)
	return
}
