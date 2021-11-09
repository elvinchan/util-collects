package log

import (
	"fmt"
	"testing"
)

type testReceiver struct {
	outputer func(entry *BasicEntry, lvl Level, msg string)
}

func (r *testReceiver) Output(entry *BasicEntry, lvl Level, msg string) {
	r.outputer(entry, lvl, msg)
}

func TestCustomizedReceiver(t *testing.T) {
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
			l := NewLogger(c.Prefix, &testReceiver{
				outputer: func(entry *BasicEntry, lvl Level, msg string) {
					resultCh <- fmt.Sprintf("%s %s - %s", lvl, entry.Logger.Prefix, msg)
					fieldsCh <- entry.Fields
				},
			})
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
			for k, v := range c.Fields {
				e.WithField(k, v)
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
			fields = <-fieldsCh
			if len(fields) != 0 {
				t.Errorf("fields length not match, expect %d, got %d", 0, len(fields))
			}
		})
	}
}
