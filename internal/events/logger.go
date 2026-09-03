package events

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/rs/zerolog"
)

// zerologAdapter bridges Watermill's LoggerAdapter onto the zerolog logger
// used everywhere else in the application, so broker logs land in the same
// stream and format as the HTTP and database logs.
type zerologAdapter struct {
	logger zerolog.Logger
}

// NewWatermillLogger wraps a zerolog logger for Watermill.
func NewWatermillLogger(logger zerolog.Logger) watermill.LoggerAdapter {
	return &zerologAdapter{logger: logger}
}

func (a *zerologAdapter) Error(msg string, err error, fields watermill.LogFields) {
	a.logger.Error().Err(err).Fields(map[string]any(fields)).Msg(msg)
}

func (a *zerologAdapter) Info(msg string, fields watermill.LogFields) {
	a.logger.Info().Fields(map[string]any(fields)).Msg(msg)
}

func (a *zerologAdapter) Debug(msg string, fields watermill.LogFields) {
	a.logger.Debug().Fields(map[string]any(fields)).Msg(msg)
}

func (a *zerologAdapter) Trace(msg string, fields watermill.LogFields) {
	a.logger.Trace().Fields(map[string]any(fields)).Msg(msg)
}

func (a *zerologAdapter) With(fields watermill.LogFields) watermill.LoggerAdapter {
	return &zerologAdapter{
		logger: a.logger.With().Fields(map[string]any(fields)).Logger(),
	}
}
