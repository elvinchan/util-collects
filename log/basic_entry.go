package log

import (
	"context"
	"fmt"
	"os"
)

type BasicEntry struct {
	Logger *BasicLogger
	Fields map[string]interface{}
	Ctx    context.Context
}

func (e *BasicEntry) WithContext(ctx context.Context) Entry {
	return &BasicEntry{Logger: e.Logger, Fields: e.Fields, Ctx: ctx}
}

func (e *BasicEntry) WithField(key string, value interface{}) Entry {
	return e.WithFields(map[string]interface{}{key: value})
}

func (e *BasicEntry) WithFields(fields Fields) Entry {
	data := make(map[string]interface{}, len(e.Fields)+len(fields))
	for k, v := range e.Fields {
		data[k] = v
	}
	for k, v := range fields {
		data[k] = v
	}
	return &BasicEntry{Logger: e.Logger, Fields: data, Ctx: e.Ctx}
}

func (e *BasicEntry) WithError(err error) Entry {
	return e.WithField(ErrorKey, err)
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
