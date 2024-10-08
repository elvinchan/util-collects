package log

import "context"

func NewDiscardLogger() Logger {
	return discardLogger{}
}

type discardLogger struct{}

func (d discardLogger) NewEntry() Entry                               { return d }
func (d discardLogger) AddPrefix(p string) Entry                      { return d }
func (d discardLogger) WithContext(ctx context.Context) Entry         { return d }
func (d discardLogger) WithField(key string, value interface{}) Entry { return d }
func (d discardLogger) WithFields(keysAndValues ...interface{}) Entry { return d }
func (d discardLogger) WithError(err error) Entry                     { return d }
func (discardLogger) Fatal(format string, v ...interface{})           {}
func (discardLogger) Panic(format string, v ...interface{})           {}
func (discardLogger) Error(format string, v ...interface{})           {}
func (discardLogger) Warn(format string, v ...interface{})            {}
func (discardLogger) Info(format string, v ...interface{})            {}
func (discardLogger) Debug(format string, v ...interface{})           {}
func (discardLogger) Level() Level                                    { return 0 }
func (discardLogger) SetLevel(lvl Level)                              {}
