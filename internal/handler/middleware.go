package handler

import (
	"assistant-go/internal/locale"
	"context"
	"net/http"
)

// Middleware константы
const (
	LocaleMW = "LocaleMW"
	AuthMW   = "AuthMW"
)

var MapMiddleware = map[string]Middleware{
	LocaleMW: LocaleMiddleware,
	AuthMW:   AuthMiddleware,
}

type Middleware func(http.HandlerFunc) http.HandlerFunc

func ApplyMiddleware(h http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	for i := 0; i < len(middlewares); i++ {
		h = middlewares[i](h)
	}
	return h
}

func BuildHandler(h http.HandlerFunc, middlewares ...string) http.HandlerFunc {
	var stack []Middleware
	for _, mw := range middlewares {
		if middlewareFunc, exists := MapMiddleware[mw]; exists {
			stack = append(stack, middlewareFunc)
		}
	}
	return ApplyMiddleware(h, stack...)
}

func LocaleMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return locale.Middleware(next)
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		r = r.WithContext(context.WithValue(r.Context(), "user", 1))
		next.ServeHTTP(w, r)
	}
	//return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//	token := r.Header.Get("Authorization")
	//	if token == "" {
	//		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	//		return
	//	}
	//	// Здесь логика проверки токена и извлечения пользователя
	//	// user := parseToken(token)
	//	// r = r.WithContext(context.WithValue(r.Context(), "user", user))
	//	next.ServeHTTP(w, r)
	//})
}
