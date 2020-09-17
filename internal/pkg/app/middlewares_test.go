package app

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	testLog "github.com/xmidt-org/ears/test/log"
	"net/http"
	"net/http/httptest"
	"testing"
)

type Validator struct {
	validate func(w http.ResponseWriter, r *http.Request)
}

func (v *Validator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v.validate(w, r)
}

func TestInitRequestMiddleware(t *testing.T) {

	listener := testLog.NewLogListener()
	logger := zerolog.New(listener)
	middlewares := NewMiddlewares(&logger)

	//Test Case 1
	validator := &Validator{func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log.Ctx(ctx).Debug().Msg("test")
		listener.AssertLastLogLine(t, "gears.app.id", "myapp")
		listener.AssertLastLogLine(t, "tx.traceId", "123456")
	}}

	middleware := middlewares[0](validator)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Application-Id", "myapp")
	r.Header.Set("X-B3-TraceId", "123456")

	middleware.ServeHTTP(w, r)

	//Test Case 2
	validator = &Validator{func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log.Ctx(ctx).Debug().Msg("test")
		listener.AssertLastLogLine(t, "gears.app.id", "myapp2")
	}}
	middleware = middlewares[0](validator)
	r = httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Application-Id", "myapp2")

	middleware.ServeHTTP(w, r)
}

func TestAuthMiddleware(t *testing.T) {
	ctx := context.Background()
	listener := testLog.NewLogListener()
	logger := zerolog.New(listener)
	middlewares := NewMiddlewares(&logger)
	subCtx := logger.WithContext(ctx)

	//AuthenticateMiddleware is currently just a pass through.
	//Just validate that it actually goes through
	validator := &Validator{func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log.Ctx(ctx).Debug().Msg("good")
		listener.AssertLastLogLine(t, "message", "good")
	}}

	middleware := middlewares[1](validator)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Application-Id", "myapp")
	r.Header.Set("X-B3-TraceId", "123456")
	//r.WithContext(log.Logger.WithContext(context.Background()))

	middleware.ServeHTTP(w, r.WithContext(subCtx))
}
