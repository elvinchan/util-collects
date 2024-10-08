package log

import (
	"context"
	"fmt"
	"testing"
)

type testSink struct {
	fields   map[string]interface{}
	outputer func(sink *testSink, prefix string, lvl Level, msg string)
}

func (s *testSink) WithField(key string, value interface{}) Sink {
	newMap := make(map[string]interface{}, len(s.fields)+1)
	for k, v := range s.fields {
		newMap[k] = v
	}
	newMap[key] = value

	clone := *s
	clone.fields = newMap
	return &clone
}

func (s *testSink) Output(ctx context.Context, prefix string, lvl Level, msg string) {
	s.outputer(s, prefix, lvl, msg)
}

func TestCustomizedSink(t *testing.T) {
	cases := []struct {
		Prefix string
		Level  Level
		Fields map[string]interface{}
		Format string
		Args   []interface{}
		Expect string
	}{
		{
			Prefix: "hello",
			Level:  InfoLevel,
			Fields: map[string]interface{}{
				"int":    1,
				"array":  []int{2, 3},
				"string": "4",
				"bool":   false,
			},
			Format: "log expect balabala",
			Args:   nil,
			Expect: "INFO hello - log expect balabala",
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("Case-%d", i), func(t *testing.T) {
			resultCh := make(chan string, 1)
			fieldsCh := make(chan map[string]interface{}, 1)
			sink := &testSink{}
			sink.outputer = func(sink *testSink, prefix string, lvl Level, msg string) {
				fieldsCh <- sink.fields
				resultCh <- fmt.Sprintf("%s %s - %s", lvl, prefix, msg)
			}
			l := NewBasicLogger(sink, BasicLoggerWithPrefix(c.Prefix))
			l.SetLevel(DebugLevel)
			e := l.NewEntry()
			for k, v := range c.Fields {
				e = e.WithField(k, v)
			}

			switch c.Level {
			case DebugLevel:
				e.Debug(c.Format, c.Args...)
			case InfoLevel:
				e.Info(c.Format, c.Args...)
			case WarnLevel:
				e.Warn(c.Format, c.Args...)
			case ErrorLevel:
				e.Error(c.Format, c.Args...)
			}
			if msg := <-resultCh; msg != c.Expect {
				t.Errorf("expect %s, got %s", c.Expect, msg)
			}
			fields := <-fieldsCh
			if len(fields) != len(c.Fields) {
				t.Errorf("fields length not match, expect %d, got %d", len(c.Fields), len(fields))
			}

			e = l.NewEntry()

			switch c.Level {
			case DebugLevel:
				e.Debug(c.Format, c.Args...)
			case InfoLevel:
				e.Info(c.Format, c.Args...)
			case WarnLevel:
				e.Warn(c.Format, c.Args...)
			case ErrorLevel:
				e.Error(c.Format, c.Args...)
			}
			if msg := <-resultCh; msg != c.Expect {
				t.Errorf("expect %s, got %s", c.Expect, msg)
			}
			fields = <-fieldsCh
			if len(fields) != 0 {
				t.Errorf("fields length not match, expect %d, got %d", 0, len(fields))
			}
		})
	}
}

func TestDiscard(t *testing.T) {
	l := NewDiscardLogger()
	l.Debug("d")
	l.Info("d")
	l.Warn("d")
	l.Error("d")
	l.Panic("d")
	l.Fatal("d")

	l.WithContext(context.Background()).WithField("k", "v").
		WithError(nil).Debug("ds")
}
