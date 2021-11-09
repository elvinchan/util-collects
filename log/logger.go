package log

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
)

type (
	Logger interface {
		NewEntry() Entry
		Entry

		Level() Level
		SetLevel(lvl Level)
	}

	BasicLogger struct {
		Prefix   string
		level    uint32
		receiver Receiver
	}

	// Level type
	Level uint32
)

const (
	// FatalLevel level. Logs and then calls `os.Exit(1)`.
	FatalLevel Level = iota
	// PanicLevel level. Logs and then calls panic with the message passed to
	// Debug, Info, ...
	PanicLevel
	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	ErrorLevel
	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel
	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	InfoLevel
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel
)

// Defines the key when adding errors using WithError.
var ErrorKey = "error"

func (lvl Level) String() string {
	if lvl == FatalLevel {
		return "FATAL"
	} else if lvl == PanicLevel {
		return "PANIC"
	} else if lvl == ErrorLevel {
		return "ERROR"
	} else if lvl == WarnLevel {
		return "WARN"
	} else if lvl == InfoLevel {
		return "INFO"
	} else if lvl == DebugLevel {
		return "DEBUG"
	} else {
		return ""
	}
}

const defaultSeparator = " | "

func NewDefaultLogger(prefix string) Logger {
	return NewLogger(prefix, newDefaultReceiver())
}

func NewLogger(prefix string, receiver Receiver) Logger {
	return &BasicLogger{
		Prefix:   prefix,
		level:    uint32(InfoLevel),
		receiver: receiver,
	}
}

func (l *BasicLogger) NewEntry() Entry {
	return &BasicEntry{
		Logger: l,
		Fields: make(map[string]interface{}),
	}
}

func (l *BasicLogger) WithContext(ctx context.Context) Entry {
	return &BasicEntry{
		Logger: l,
		Fields: make(map[string]interface{}),
		Ctx:    ctx,
	}
}

func (l *BasicLogger) WithField(key string, value interface{}) Entry {
	return &BasicEntry{
		Logger: l,
		Fields: map[string]interface{}{
			key: value,
		},
	}
}

func (l *BasicLogger) WithFields(fields Fields) Entry {
	if fields == nil {
		return l.NewEntry()
	}
	return &BasicEntry{
		Logger: l,
		Fields: fields,
	}
}

func (l *BasicLogger) WithError(err error) Entry {
	return &BasicEntry{
		Logger: l,
		Fields: map[string]interface{}{
			ErrorKey: err,
		},
	}
}

func (l *BasicLogger) Fatal(format string, v ...interface{}) {
	entry := &BasicEntry{Logger: l}
	entry.log(FatalLevel, format, v...)
	os.Exit(1)
}

func (l *BasicLogger) Panic(format string, v ...interface{}) {
	entry := &BasicEntry{Logger: l}
	entry.log(PanicLevel, format, v...)
	panic(fmt.Sprintf(format, v...))
}

func (l *BasicLogger) Error(format string, v ...interface{}) {
	entry := &BasicEntry{Logger: l}
	entry.log(ErrorLevel, format, v...)
}

func (l *BasicLogger) Warn(format string, v ...interface{}) {
	entry := &BasicEntry{Logger: l}
	entry.log(WarnLevel, format, v...)
}

func (l *BasicLogger) Info(format string, v ...interface{}) {
	entry := &BasicEntry{Logger: l}
	entry.log(InfoLevel, format, v...)
}

func (l *BasicLogger) Debug(format string, v ...interface{}) {
	entry := &BasicEntry{Logger: l}
	entry.log(DebugLevel, format, v...)
}

func (l *BasicLogger) SetLevel(level Level) {
	atomic.StoreUint32(&l.level, uint32(level))
}

func (l *BasicLogger) Level() Level {
	return Level(atomic.LoadUint32(&l.level))
}
