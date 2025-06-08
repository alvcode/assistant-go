package handler

import (
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/layer/ucase"
	"assistant-go/internal/locale"
	"encoding/json"
	"fmt"
	"net/http"
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

	err = h.useCase.CreateDirectory(&createDirectoryDTO, authUser)
	if err != nil {
		SendErrorResponse(w, buildErrorMessage(langRequest, err), http.StatusUnprocessableEntity, 0)
		return
	}
	SendResponse(w, http.StatusCreated, "HELLO")
	return
}
