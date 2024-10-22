package utils

import (
	"errors"

	"github.com/rs/zerolog"
)

var (
	ErrUUIDTooLong         = errors.New("requested uuid is too long")
	ErrUUIDGeneration      = errors.New("failed to generate uuid")
	ErrInvalidZerologLevel = errors.New("invalid zerolog level")
)

// SetZeroLogLevel sets the global log level for zerolog.
func SetZerologLevel(level string) error {
	switch level {
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	default:
		return ErrInvalidZerologLevel
	}
	return nil
}
