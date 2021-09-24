package log

import (
	"context"
	"fmt"
	"os"
)

type (
	Fields map[string]interface{}

	Entry interface {
		WithContext(ctx context.Context) Entry
		WithField(key string, value interface{}) Entry
		WithFields(fields Fields) Entry
		WithError(err error) Entry

		Fatal(format string, v ...interface{})
		Panic(format string, v ...interface{})
		Error(format string, v ...interface{})
		Warn(format string, v ...interface{})
		Info(format string, v ...interface{})
		Debug(format string, v ...interface{})
	}

	BasicEntry struct {
		Logger *BasicLogger
		Fields map[string]interface{}
		Ctx    context.Context
	}
)

func (e *BasicEntry) WithContext(ctx context.Context) Entry {
	e.Ctx = ctx
	return e
}

func (e *BasicEntry) WithField(key string, value interface{}) Entry {
	e.Fields[key] = value
	return e
}

func (e *BasicEntry) WithFields(fields Fields) Entry {
	e.Fields = fields
	return e
}

func (e *BasicEntry) WithError(err error) Entry {
	e.Fields[ErrorKey] = err.Error()
	return e
}

func (e *BasicEntry) Fatal(format string, v ...interface{}) {
	e.log(FatalLevel, format, v...)
	os.Exit(1)
}

func (e *BasicEntry) Panic(format string, v ...interface{}) {
	e.log(PanicLevel, format, v...)
	panic(fmt.Sprintf(format, v...))
}

func (e *BasicEntry) Error(format string, v ...interface{}) {
	e.log(ErrorLevel, format, v...)
}

func (e *BasicEntry) Warn(format string, v ...interface{}) {
	e.log(WarnLevel, format, v...)
}

func (e *BasicEntry) Info(format string, v ...interface{}) {
	e.log(InfoLevel, format, v...)
}

func (e *BasicEntry) Debug(format string, v ...interface{}) {
	e.log(DebugLevel, format, v...)
}

func (e *BasicEntry) log(lvl Level, format string, v ...interface{}) {
	if lvl > e.Logger.Level() {
		return
	}
	e.Logger.receiver.Output(e, lvl, fmt.Sprintf(format, v...))
}
