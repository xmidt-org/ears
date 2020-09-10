package app

import (
	"context"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
)

var middlewareLogger *zerolog.Logger

func NewMiddlewares(logger *zerolog.Logger) []func(next http.Handler) http.Handler {
	middlewareLogger = logger
	return []func(next http.Handler) http.Handler{
		initRequestMiddleware,
		authenticateMiddleware,
	}
}

func getSubCtxWithStr(ctx context.Context, key string, value string) context.Context {
	//TODO this helper function is lame because my go compiler is not happy when WithContext is chain to the end of the first call
	subLogger := middlewareLogger.With().Str(key, value).Logger()
	return subLogger.WithContext(ctx)
}

func initRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		traceId := r.Header.Get(HeaderTraceId)
		if traceId == "" {
			traceId = uuid.New().String()
		}
		subCtx := getSubCtxWithStr(ctx, LogTraceId, traceId)

		appId := r.Header.Get(HeaderTenantId)
		if appId != "" {
			subCtx = getSubCtxWithStr(ctx, LogTenantId, appId)
		}

		log.Ctx(subCtx).Debug().Msg("initializeRequestMiddleware")

		next.ServeHTTP(w, r.WithContext(subCtx))
	})
}

func authenticateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log.Ctx(ctx).Debug().Msg("authenticateMiddleare")
		next.ServeHTTP(w, r)
	})
}
