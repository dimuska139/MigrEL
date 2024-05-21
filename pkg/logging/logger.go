package logging

import (
	"fmt"
	"github.com/rs/zerolog"
	"time"
)

type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

type Logger struct {
	logger zerolog.Logger
}

func NewLogger(loglevel LogLevel) *Logger {
	logger := zerolog.New(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.TimeFormat = time.RFC3339
	})).
		With().
		Timestamp().
		Logger()

	if loglevel != "" {
		switch loglevel {
		case LogLevelWarn:
			zerolog.SetGlobalLevel(zerolog.WarnLevel)
		case LogLevelInfo:
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		case LogLevelError:
			zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		case LogLevelDebug:
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		default:
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		}
	}

	return &Logger{logger}
}

func (z Logger) Debug(msg string, args ...any) {
	z.logger.Debug().Timestamp().Fields(args).Msg(msg)
}

func (z Logger) Info(msg string, args ...any) {
	z.logger.Info().Timestamp().Fields(args).Msg(msg)
}

func (z Logger) Warn(msg string, args ...any) {
	z.logger.Warn().Timestamp().Fields(args).Msg(msg)
}

func (z Logger) Error(msg string, args ...any) {
	z.logger.Error().Timestamp().Fields(args).Msg(msg)
}

func (z Logger) Panic(msg string, args ...any) {
	z.logger.Panic().Timestamp().Fields(args).Msg(msg)
}

func (z Logger) Printf(format string, v ...any) {
	z.logger.Printf(format, v...)
}

func (z Logger) Fatal(v ...any) {
	z.logger.Fatal().Timestamp().Msg(fmt.Sprint(v...))
}

func (z Logger) Fatalf(format string, args ...any) {
	z.logger.Fatal().Timestamp().Msgf(format, args...)
}

func (z Logger) Println(args ...any) {
	z.logger.Info().Timestamp().Msgf("%v\r\n", args...)
}

func (z Logger) Print(args ...any) {
	z.logger.Print(args...)
}
