package log

import (
	"context"
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
)
