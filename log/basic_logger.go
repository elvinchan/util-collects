package log

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
)

type BasicLogger struct {
	Prefix   string
	level    uint32
	receiver Receiver
}

const basicSeparator = " | "

func NewDefaultBasicLogger(prefix string) Logger {
	return NewBasicLogger(prefix, defaultReceiver())
}

func NewBasicLogger(prefix string, receiver Receiver) Logger {
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
