package bin

import (
	"fmt"

	"go.uber.org/zap"
)

var loggingEnabled = false

var encoderLog = zap.NewNop()
var decoderLog = zap.NewNop()

type logStringerFunc func() string

func (f logStringerFunc) String() string { return f() }

func typeField(field string, v interface{}) zap.Field {
	return zap.Stringer(field, logStringerFunc(func() string {
		return fmt.Sprintf("%T", v)
	}))
}

func newLogger(production bool) (l *zap.Logger) {
	if production {
		l, _ = zap.NewProduction()
	} else {
		l, _ = zap.NewDevelopment()
	}
	return
}

// NewLogger a wrap to newLogger
func NewLogger(production bool) *zap.Logger {
	return newLogger(production)
}

func EnableDebugLogging(l *zap.Logger) {
	encoderLog = l
	decoderLog = l
	loggingEnabled = true
}
