package log

import (
	"context"
	"encoding/json"
	stdlog "log"
	"os"
	"strings"
)

type basicSink struct {
	logger *stdlog.Logger
	fields map[string]interface{} // TODO: use ordered map for stable of keys
}

func (s *basicSink) WithField(key string, value interface{}) Sink {
	newMap := make(map[string]interface{}, len(s.fields)+1)
	for k, v := range s.fields {
		newMap[k] = v
	}
	newMap[key] = value

	clone := *s
	clone.fields = newMap
	return &clone
}

const basicSeparator = " | "

func (s *basicSink) Output(ctx context.Context, prefix string, lvl Level, msg string) {
	var sb strings.Builder
	sb.WriteString(lvl.String())
	sb.WriteString(basicSeparator)
	sb.WriteString(prefix)
	if len(s.fields) > 0 {
		fs, err := json.Marshal(s.fields)
		if err == nil {
			sb.WriteString(basicSeparator)
			sb.Write(fs)
		}
	}
	sb.WriteString(basicSeparator)
	sb.WriteString(msg)
	s.logger.Println(sb.String())
}

func defaultSink() Sink {
	ol := stdlog.New(os.Stderr, "", stdlog.LstdFlags)
	return &basicSink{
		logger: ol,
	}
}
