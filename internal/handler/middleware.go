package handler

import (
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/layer/entity"
	"assistant-go/internal/locale"
	"assistant-go/internal/logging"
	"assistant-go/internal/storage/postgres"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const UserContextKey = "user"

const (
	LocaleMW      = "LocaleMW"
	AuthMW        = "AuthMW"
	BlockIPMW     = "BlockIPMW"
	RateLimiterMW = "RateLimiterMW"
)

var MapMiddleware = map[string]Middleware{
	LocaleMW:      LocaleMiddleware,
	AuthMW:        AuthMiddleware,
	BlockIPMW:     BlockIPMiddleware,
	RateLimiterMW: RateLimiterMiddleware,
}

type Middleware func(http.HandlerFunc) http.HandlerFunc

func ApplyMiddleware(h http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

func BuildHandler(h http.HandlerFunc, middlewares ...string) http.HandlerFunc {
	stack := []Middleware{
		MapMiddleware[BlockIPMW],
		MapMiddleware[RateLimiterMW],
		MapMiddleware[LocaleMW],
	}

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

		userTokenEntity, err := userRepository.FindUserToken(r.Context(), dtoUserToken.Token)
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

		userEntity, err := userRepository.FindById(r.Context(), userTokenEntity.UserId)
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

func RateLimiterMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		langRequest := locale.GetLangFromContext(r.Context())

		allowRequests := appConf.RateLimiter.AllowanceRequests
		timeDuration := appConf.RateLimiter.TimeDurationSec

		IPAddress, err := GetIpAddress(r)
		if err != nil {
			SendErrorResponse(w, buildErrorMessage(langRequest, ErrSplitHostIP), http.StatusForbidden, 0)
			return
		}

		exists := true
		foundIP, err := rateLimiterRepository.FindIP(IPAddress)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				exists = false
			} else {
				logging.GetLogger(r.Context()).Error(err)
				SendErrorResponse(w, buildErrorMessage(langRequest, postgres.ErrUnexpectedDBError), http.StatusUnprocessableEntity, 0)
				return
			}
		}

		limiter := &entity.RateLimiter{
			IP:                IPAddress,
			AllowanceRequests: allowRequests,
			Timestamp:         time.Now().Unix(),
		}

		// записи нет, создаем
		if exists == false {
			err := rateLimiterRepository.UpsertIP(limiter)
			if err != nil {
				logging.GetLogger(r.Context()).Error(err)
				SendErrorResponse(w, buildErrorMessage(langRequest, postgres.ErrUnexpectedDBError), http.StatusUnprocessableEntity, 0)
				return
			}
		}

		// диапазон проверки протух. сброс !до настроечных!
		limiterExpired := false
		if exists && (time.Now().Unix()-foundIP.Timestamp > int64(timeDuration)) {
			limiterExpired = true
			err := rateLimiterRepository.UpsertIP(limiter)
			if err != nil {
				logging.GetLogger(r.Context()).Error(err)
				SendErrorResponse(w, buildErrorMessage(langRequest, postgres.ErrUnexpectedDBError), http.StatusUnprocessableEntity, 0)
				return
			}
		}

		var updateEntity *entity.RateLimiter
		if !exists || (exists && limiterExpired) {
			updateEntity = &entity.RateLimiter{
				IP:                limiter.IP,
				AllowanceRequests: limiter.AllowanceRequests,
				Timestamp:         limiter.Timestamp,
			}
		} else if exists && !limiterExpired {
			updateEntity = &entity.RateLimiter{
				IP:                foundIP.IP,
				AllowanceRequests: foundIP.AllowanceRequests,
				Timestamp:         foundIP.Timestamp,
			}
		}

		if updateEntity.AllowanceRequests > 0 {
			updateEntity, err = rateLimiterRepository.UpdateIP(updateEntity)
			if err != nil {
				logging.GetLogger(r.Context()).Error(err)
				SendErrorResponse(w, buildErrorMessage(langRequest, postgres.ErrUnexpectedDBError), http.StatusUnprocessableEntity, 0)
				return
			}
			w.Header().Set("X-Rate-Limit-Limit", strconv.Itoa(allowRequests))
			w.Header().Set("X-Rate-Limit-Remaining", strconv.Itoa(updateEntity.AllowanceRequests))
		}

		if updateEntity.AllowanceRequests <= 0 {
			BlockEventHandle(r, BlockEventTooManyRequests)
			SendErrorResponse(w, "Too Many Requests", http.StatusTooManyRequests, 0)
			return
		}

		next(w, r)
	}
}
