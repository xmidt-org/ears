package logs

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

//Helper function to derive context from a parent context in zerolog
func SubLoggerCtx(ctx context.Context, parent *zerolog.Logger) context.Context {
	subCtx := parent.WithContext(ctx)
	logger := log.Ctx(subCtx).With().Logger()
	return logger.WithContext(subCtx)
}

//Helper function to add zerolog key/value pari to a golang context
func StrToLogCtx(ctx context.Context, key string, value string) {
	l := zerolog.Ctx(ctx)
	l.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str(key, value)
	})
}
