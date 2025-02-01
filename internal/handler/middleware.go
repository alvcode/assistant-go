package handler

import (
	"assistant-go/internal/locale"
	"net/http"
)

func LocaleMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return locale.Middleware(next)
}
