package handler

import (
	"assistant-go/internal/layer/ucase"
	"assistant-go/internal/locale"
	"fmt"
	"net/http"
)

type DriveRecycleBinHandler struct {
	useCase ucase.DriveRecycleBinUseCase
}

func NewDriveRecycleBinHandler(useCase ucase.DriveRecycleBinUseCase) *DriveRecycleBinHandler {
	return &DriveRecycleBinHandler{
		useCase: useCase,
	}
}

func (h *DriveRecycleBinHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	langRequest := locale.GetLangFromContext(r.Context())
	authUser, err := GetAuthUser(r)
	if err != nil {
		BlockEventHandle(r, BlockEventUnauthorizedType)
		SendErrorResponse(w, locale.T(langRequest, "unauthorized"), http.StatusUnauthorized, 0)
		return
	}

	fmt.Println(authUser.ID)
}
