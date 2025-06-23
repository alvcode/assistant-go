package handler

import (
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/layer/repository"
	"assistant-go/internal/layer/ucase"
	"assistant-go/internal/layer/vmodel"
	"assistant-go/internal/locale"
	"assistant-go/internal/logging"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
)

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
		structIDInt, err := strconv.Atoi(structIDStr)

		if err != nil {
			BlockEventHandle(r, BlockEventInputDataType)
			SendErrorResponse(w, locale.T(langRequest, "parameter_conversion_error"), http.StatusBadRequest, 0)
			return
		}
		structID = structIDInt
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

func (h *DriveHandler) GetFile(w http.ResponseWriter, r *http.Request) {
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
		structIDInt, err := strconv.Atoi(structIDStr)

		if err != nil {
			BlockEventHandle(r, BlockEventInputDataType)
			SendErrorResponse(w, locale.T(langRequest, "parameter_conversion_error"), http.StatusBadRequest, 0)
			return
		}
		structID = structIDInt
	}

	fileDto, err := h.useCase.GetFile(structID, appConf.Drive.SavePath, authUser)
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

func (h *DriveHandler) Rename(w http.ResponseWriter, r *http.Request) {
	langRequest := locale.GetLangFromContext(r.Context())
	var renameDTO dto.DriveRenameStruct

	authUser, err := GetAuthUser(r)
	if err != nil {
		BlockEventHandle(r, BlockEventUnauthorizedType)
		SendErrorResponse(w, locale.T(langRequest, "unauthorized"), http.StatusUnauthorized, 0)
		return
	}

	var structID int
	params := httprouter.ParamsFromContext(r.Context())
	if structIDStr := params.ByName("id"); structIDStr != "" {
		structIDInt, err := strconv.Atoi(structIDStr)

		if err != nil {
			BlockEventHandle(r, BlockEventInputDataType)
			SendErrorResponse(w, locale.T(langRequest, "parameter_conversion_error"), http.StatusBadRequest, 0)
			return
		}
		structID = structIDInt
	}

	err = json.NewDecoder(r.Body).Decode(&renameDTO)
	if err != nil {
		BlockEventHandle(r, BlockEventDecodeBodyType)
		SendErrorResponse(w, locale.T(langRequest, "error_reading_request_body"), http.StatusBadRequest, 0)
		return
	}

	if err = renameDTO.Validate(langRequest); err != nil {
		BlockEventHandle(r, BlockEventInputDataType)
		SendErrorResponse(w, fmt.Sprint(err), http.StatusUnprocessableEntity, 0)
		return
	}

	err = h.useCase.Rename(structID, renameDTO.Name, authUser)
	if err != nil {
		SendErrorResponse(w, buildErrorMessage(langRequest, err), http.StatusUnprocessableEntity, 0)
		return
	}

	SendResponse(w, http.StatusNoContent, nil)
	return
}

func (h *DriveHandler) Space(w http.ResponseWriter, r *http.Request) {
	langRequest := locale.GetLangFromContext(r.Context())

	authUser, err := GetAuthUser(r)
	if err != nil {
		BlockEventHandle(r, BlockEventUnauthorizedType)
		SendErrorResponse(w, locale.T(langRequest, "unauthorized"), http.StatusUnauthorized, 0)
		return
	}

	dtoSpace, err := h.useCase.Space(authUser, appConf.Drive.LimitPerUser<<20)
	if err != nil {
		SendErrorResponse(w, buildErrorMessage(langRequest, err), http.StatusUnprocessableEntity, 0)
		return
	}

	SendResponse(w, http.StatusOK, dtoSpace)
	return
}
