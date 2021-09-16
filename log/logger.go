package log

import (
	"encoding/json"
	"fmt"
	stdlog "log"
	"os"
	"sync/atomic"
)

type (
	Output func(i GetInfo, level Level, fields map[string]interface{},
		msg string)

	Logger interface {
		NewEntry() Entry
		Entry

		GetInfo
		SetInfo
	}

	GetInfo interface {
		Prefix() string
		Level() Level
	}

	SetInfo interface {
		SetPrefix(p string)
		SetLevel(level Level)
	}

	logger struct {
		prefix string
		level  uint32
		output Output
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

func (level Level) String() string {
	if level == FatalLevel {
		return "FATAL"
	} else if level == PanicLevel {
		return "PANIC"
	} else if level == ErrorLevel {
		return "ERROR"
	} else if level == WarnLevel {
		return "WARN"
	} else if level == InfoLevel {
		return "INFO"
	} else if level == DebugLevel {
		return "DEBUG"
	} else {
		return ""
	}
}

func NewDefaultLogger(prefix string) *logger {
	ol := stdlog.New(os.Stderr, "", stdlog.LstdFlags)
	return NewLogger(prefix, Output(func(i GetInfo, level Level,
		fields map[string]interface{}, msg string) {
		fs, _ := json.Marshal(fields)
		if i.Prefix() == "" {
			ol.Printf("%s %s | %s\n", level, fs, msg)
		} else {
			ol.Printf("%s %s | %s | %s\n", level, i.Prefix(), fs, msg)
		}
	}))
}

func NewLogger(prefix string, output Output) *logger {
	return &logger{
		prefix: prefix,
		level:  uint32(InfoLevel),
		output: output,
	}
}

func (l *logger) NewEntry() Entry {
	return &entry{logger: l, fields: make(map[string]interface{})}
}

func (l *logger) WithField(key string, value interface{}) Entry {
	return &entry{
		logger: l,
		fields: map[string]interface{}{
			key: value,
		},
	}
}

func (l *logger) WithError(err error) Entry {
	return &entry{
		logger: l,
		fields: map[string]interface{}{
			ErrorKey: err,
		},
	}
}

func (l *logger) Fatal(format string, v ...interface{}) {
	l.log(FatalLevel, format, v...)
	os.Exit(1)
}

func (l *logger) Panic(format string, v ...interface{}) {
	l.log(PanicLevel, format, v...)
	panic(fmt.Sprintf(format, v...))
}

func (l *logger) Error(format string, v ...interface{}) {
	l.log(ErrorLevel, format, v...)
}

func (l *logger) Warn(format string, v ...interface{}) {
	l.log(WarnLevel, format, v...)
}

func (l *logger) Info(format string, v ...interface{}) {
	l.log(InfoLevel, format, v...)
}

func (l *logger) Debug(format string, v ...interface{}) {
	l.log(DebugLevel, format, v...)
}

func (l *logger) Prefix() string {
	return l.prefix
}

func (l *logger) SetPrefix(p string) {
	l.prefix = p
}

func (l *logger) SetLevel(level Level) {
	atomic.StoreUint32(&l.level, uint32(level))
}

func (l *logger) Level() Level {
	return Level(atomic.LoadUint32(&l.level))
}

func (l *logger) log(level Level, format string, v ...interface{}) {
	l.logWithFields(level, nil, fmt.Sprintf(format, v...))
}

func (l *logger) logWithFields(level Level, fields map[string]interface{},
	format string, v ...interface{}) {
	if level > l.Level() {
		return
	}
	l.output(l, level, fields, fmt.Sprintf(format, v...))
}
