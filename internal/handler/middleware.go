package handler

import (
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/locale"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"net/http"
	"strings"
	"time"
)

const UserContextKey = "user"

const (
	LocaleMW  = "LocaleMW"
	AuthMW    = "AuthMW"
	BlockIPMW = "BlockIPMW"
)

var MapMiddleware = map[string]Middleware{
	LocaleMW:  LocaleMiddleware,
	AuthMW:    AuthMiddleware,
	BlockIPMW: BlockIPMiddleware,
}

type Middleware func(http.HandlerFunc) http.HandlerFunc

func ApplyMiddleware(h http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	for i := len(middlewares) - 1; i >= 0; i-- {
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
		langRequest := locale.GetLangFromContext(r.Context())

		header := r.Header.Get("Authorization")
		if header == "" {
			BlockEventHandle(r, BlockEventUnauthorizedType)
			SendErrorResponse(w, locale.T(langRequest, "unauthorized"), http.StatusUnauthorized, 0)
			return
		}
		const prefix = "Bearer "
		if !strings.HasPrefix(header, prefix) {
			BlockEventHandle(r, BlockEventUnauthorizedType)
			SendErrorResponse(w, locale.T(langRequest, "unauthorized"), http.StatusUnauthorized, 0)
			return
		}

		token := strings.TrimPrefix(header, prefix)
		dtoUserToken := dto.UserToken{Token: token}

		if err := dtoUserToken.Validate(langRequest); err != nil {
			BlockEventHandle(r, BlockEventUnauthorizedType)
			SendErrorResponse(w, locale.T(langRequest, "unauthorized"), http.StatusUnauthorized, 0)
			return
		}

		userTokenEntity, err := userRepository.FindUserToken(dtoUserToken.Token)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				BlockEventHandle(r, BlockEventUnauthorizedType)
				SendErrorResponse(w, locale.T(langRequest, "unauthorized"), http.StatusUnauthorized, 0)
				return
			}
			BlockEventHandle(r, BlockEventUnauthorizedType)
			SendErrorResponse(w, locale.T(langRequest, "unauthorized"), http.StatusUnauthorized, 0)
			return
		}

		if userTokenEntity.ExpiredTo < int(time.Now().Unix()) {
			BlockEventHandle(r, BlockEventUnauthorizedType)
			SendErrorResponse(w, locale.T(langRequest, "unauthorized"), http.StatusUnauthorized, 0)
			return
		}

		userEntity, err := userRepository.FindById(userTokenEntity.UserId)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				BlockEventHandle(r, BlockEventUnauthorizedType)
				SendErrorResponse(w, locale.T(langRequest, "unauthorized"), http.StatusUnauthorized, 0)
				return
			}
			BlockEventHandle(r, BlockEventUnauthorizedType)
			SendErrorResponse(w, locale.T(langRequest, "unauthorized"), http.StatusUnauthorized, 0)
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, userEntity)
		next(w, r.WithContext(ctx))
	}
}

func BlockIPMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		langRequest := locale.GetLangFromContext(r.Context())

		IPAddress, err := GetIpAddress(r)
		if err != nil {
			SendErrorResponse(w, buildErrorMessage(langRequest, ErrSplitHostIP), http.StatusForbidden, 0)
			return
		}

		foundIP, err := blockIpRepository.FindBlocking(IPAddress, time.Now().UTC())
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				SendErrorResponse(w, locale.T(langRequest, "unexpected_database_error"), http.StatusForbidden, 0)
				return
			}
		}

		if foundIP == true {
			SendErrorResponse(w, locale.T(langRequest, "access_denied"), http.StatusForbidden, 0)
			return
		}

		next(w, r)
	}
}
