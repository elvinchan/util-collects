package log

import (
	"context"
	"fmt"
	"os"
)

type BasicEntry struct {
	logger *BasicLogger
	ctx    context.Context
}

func (be *BasicEntry) WithContext(ctx context.Context) Entry {
	return &BasicEntry{logger: be.logger, ctx: ctx}
}

func (be *BasicEntry) WithField(key string, value interface{}) Entry {
	return be.WithFields(key, value)
}

func (be *BasicEntry) WithFields(keysAndValues ...interface{}) Entry {
	sink := be.logger.sink
	for i := 0; i < len(keysAndValues); i += 2 {
		if i == len(keysAndValues)-1 {
			break
		}
		// process a key-value pair
		key, val := keysAndValues[i], keysAndValues[i+1]
		keyStr, isString := key.(string)
		if !isString {
			keyStr = be.nonStringKey(key)
		}
		sink = sink.WithField(keyStr, val)
	}
	return &BasicEntry{logger: be.logger.dup(sink), ctx: be.ctx}
}

func (BasicEntry) nonStringKey(v interface{}) string {
	const snipLen = 8

	snip := fmt.Sprintf("%v", v)
	if len(snip) > snipLen {
		snip = snip[:snipLen]
	}
	return fmt.Sprintf("<non-string-key: %s>", snip)
}

func (be *BasicEntry) WithError(err error) Entry {
	return be.WithField(ErrorKey, err)
}

func (be *BasicEntry) Fatal(format string, v ...interface{}) {
	be.log(FatalLevel, format, v...)
	os.Exit(1)
}

func (be *BasicEntry) Panic(format string, v ...interface{}) {
	be.log(PanicLevel, format, v...)
	panic(fmt.Sprintf(format, v...))
}

func (be *BasicEntry) Error(format string, v ...interface{}) {
	be.log(ErrorLevel, format, v...)
}

func (be *BasicEntry) Warn(format string, v ...interface{}) {
	be.log(WarnLevel, format, v...)
}

func (be *BasicEntry) Info(format string, v ...interface{}) {
	be.log(InfoLevel, format, v...)
}

func (be *BasicEntry) Debug(format string, v ...interface{}) {
	be.log(DebugLevel, format, v...)
}

func (be *BasicEntry) log(lvl Level, format string, v ...interface{}) {
	if lvl > be.logger.Level() {
		return
	}
	be.logger.sink.Output(be.ctx, be.logger.prefix, lvl, fmt.Sprintf(format, v...))
}
