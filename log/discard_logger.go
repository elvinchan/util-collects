package log

import "context"

func NewDiscardLogger() Logger {
	return discardReceiver{}
}

type discardReceiver struct{}

func (d discardReceiver) NewEntry() Entry                               { return d }
func (d discardReceiver) WithContext(ctx context.Context) Entry         { return d }
func (d discardReceiver) WithField(key string, value interface{}) Entry { return d }
func (d discardReceiver) WithFields(fields Fields) Entry                { return d }
func (d discardReceiver) WithError(err error) Entry                     { return d }
func (discardReceiver) Fatal(format string, v ...interface{})           {}
func (discardReceiver) Panic(format string, v ...interface{})           {}
func (discardReceiver) Error(format string, v ...interface{})           {}
func (discardReceiver) Warn(format string, v ...interface{})            {}
func (discardReceiver) Info(format string, v ...interface{})            {}
func (discardReceiver) Debug(format string, v ...interface{})           {}
func (discardReceiver) Level() Level                                    { return 0 }
func (discardReceiver) SetLevel(lvl Level)                              {}
