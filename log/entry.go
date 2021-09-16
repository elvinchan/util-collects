package log

import (
	"fmt"
	"os"
)

type (
	Entry interface {
		WithField(key string, value interface{}) Entry
		WithError(err error) Entry

		Debug(format string, v ...interface{})
		Info(format string, v ...interface{})
		Warn(format string, v ...interface{})
		Error(format string, v ...interface{})
	}

	entry struct {
		logger *logger
		fields map[string]interface{}
	}
)

func (e *entry) WithField(key string, value interface{}) Entry {
	e.fields[key] = value
	return e
}

func (e *entry) WithError(err error) Entry {
	e.fields[ErrorKey] = err
	return e
}

func (e *entry) Fatal(format string, v ...interface{}) {
	e.logger.logWithFields(FatalLevel, e.fields, format, v...)
	os.Exit(1)
}

func (e *entry) Panic(format string, v ...interface{}) {
	e.logger.logWithFields(PanicLevel, e.fields, format, v...)
	panic(fmt.Sprintf(format, v...))
}

func (e *entry) Error(format string, v ...interface{}) {
	e.logger.logWithFields(ErrorLevel, e.fields, format, v...)
}

func (e *entry) Warn(format string, v ...interface{}) {
	e.logger.logWithFields(WarnLevel, e.fields, format, v...)
}

func (e *entry) Info(format string, v ...interface{}) {
	e.logger.logWithFields(InfoLevel, e.fields, format, v...)
}

func (e *entry) Debug(format string, v ...interface{}) {
	e.logger.logWithFields(DebugLevel, e.fields, format, v...)
}
