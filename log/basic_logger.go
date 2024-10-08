package log

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
)

type BasicLogger struct {
	sink   Sink
	level  uint32
	prefix string
}

func NewDefaultBasicLogger(opts ...BasicLoggerOption) Logger {
	return NewBasicLogger(defaultSink(), opts...)
}

func NewBasicLogger(sink Sink, opts ...BasicLoggerOption) Logger {
	bl := &BasicLogger{
		sink:  sink,
		level: uint32(InfoLevel),
	}
	for _, opt := range opts {
		opt(bl)
	}
	return bl
}

type BasicLoggerOption func(*BasicLogger)

func BasicLoggerWithLevel(l Level) BasicLoggerOption {
	return func(bl *BasicLogger) {
		bl.level = uint32(l)
	}
}

func BasicLoggerWithPrefix(p string) BasicLoggerOption {
	return func(bl *BasicLogger) {
		bl.prefix = p
	}
}

func (bl *BasicLogger) dup(sink Sink) *BasicLogger {
	return &BasicLogger{
		sink:   sink,
		level:  atomic.LoadUint32(&bl.level),
		prefix: bl.prefix,
	}
}

func (bl *BasicLogger) NewEntry() Entry {
	return &BasicEntry{logger: bl}
}

func (bl *BasicLogger) WithContext(ctx context.Context) Entry {
	e := &BasicEntry{logger: bl}
	return e.WithContext(ctx)
}

func (bl *BasicLogger) WithField(key string, value interface{}) Entry {
	e := &BasicEntry{logger: bl}
	return e.WithField(key, value)
}

func (bl *BasicLogger) WithFields(keysAndValues ...interface{}) Entry {
	e := &BasicEntry{logger: bl}
	return e.WithFields(keysAndValues...)
}

func (bl *BasicLogger) WithError(err error) Entry {
	e := &BasicEntry{logger: bl}
	return e.WithError(err)
}

func (bl *BasicLogger) Fatal(format string, v ...interface{}) {
	entry := &BasicEntry{logger: bl}
	entry.log(FatalLevel, format, v...)
	os.Exit(1)
}

func (bl *BasicLogger) Panic(format string, v ...interface{}) {
	entry := &BasicEntry{logger: bl}
	entry.log(PanicLevel, format, v...)
	panic(fmt.Sprintf(format, v...))
}

func (bl *BasicLogger) Error(format string, v ...interface{}) {
	entry := &BasicEntry{logger: bl}
	entry.log(ErrorLevel, format, v...)
}

func (bl *BasicLogger) Warn(format string, v ...interface{}) {
	entry := &BasicEntry{logger: bl}
	entry.log(WarnLevel, format, v...)
}

func (bl *BasicLogger) Info(format string, v ...interface{}) {
	entry := &BasicEntry{logger: bl}
	entry.log(InfoLevel, format, v...)
}

func (bl *BasicLogger) Debug(format string, v ...interface{}) {
	entry := &BasicEntry{logger: bl}
	entry.log(DebugLevel, format, v...)
}

func (bl *BasicLogger) SetLevel(level Level) {
	atomic.StoreUint32(&bl.level, uint32(level))
}

func (bl *BasicLogger) Level() Level {
	return Level(atomic.LoadUint32(&bl.level))
}
