package handler

import (
	"encoding/json"
	"net/http"
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
