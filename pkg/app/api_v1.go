package app

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"
	"net/http"
	"time"
)

type APIServer struct {
	srv *http.Server
}

func startAPIServer(port int) (*APIServer, error) {
	if port < 1 {
		return nil, &InvalidOptionError{Err: fmt.Errorf("port %d cannot be less than 1", port)}
	}
	r := mux.NewRouter()
	r.HandleFunc("/version", versionHandler).Methods(http.MethodGet)

	r.Use(initializeRequestMiddleware)
	r.Use(authenticateMiddleare)

	log.Info().Str("op", "App.Run").Int("Port", port).Msg("Starting...")
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Str("op", "srv.ListenAndServe").Msg(err.Error())
		}
	}()
	return &APIServer{srv}, nil
}

func (api *APIServer) shutdown() {
	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()
	if err := api.srv.Shutdown(ctxShutDown); err != nil {
		log.Fatal().Str("op", "APIServer.shutdown").Msg(err.Error())
	}
}

func getSubCtxWithStr(ctx context.Context, key string, value string) context.Context {
	//TODO this helper function is lame because my go compiler is not happy when WithContext is chain to the end of the first call
	subLogger := log.With().Str(key, value).Logger()
	return subLogger.WithContext(ctx)
}

func initializeRequestMiddleware(next http.Handler) http.Handler {
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

func authenticateMiddleare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log.Ctx(ctx).Debug().Msg("authenticateMiddleare")
		next.ServeHTTP(w, r)
	})
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	log.Ctx(ctx).Debug().Msg("versionHandler")
	resp := Response{
		Item: Version,
	}
	resp.Respond(w)
}
