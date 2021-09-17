package log

import (
	"fmt"
	"testing"
)

type testReceiver struct {
	outputer func(i GetInfo, level Level, fields map[string]interface{},
		msg string)
}

func (r *testReceiver) Output(i GetInfo, level Level,
	fields map[string]interface{}, msg string) {
	r.outputer(i, level, fields, msg)
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
			l := NewLogger("test", &testReceiver{
				outputer: func(i GetInfo, level Level, fields map[string]interface{}, msg string) {
					resultCh <- fmt.Sprintf("%s %s - %s", level, i.Prefix(), msg)
					fieldsCh <- fields
				},
			})
			l.SetLevel(DebugLevel)
			l.SetPrefix(c.Prefix)
			e := l.NewEntry()
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
			fields := <-fieldsCh
			if len(fields) != len(c.Fields) {
				t.Errorf("fields length not match, expect %d, got %d", len(c.Fields), len(fields))
			}
		})
	}
}
