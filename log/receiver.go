package log

import (
	"encoding/json"
	stdlog "log"
	"os"
	"strings"
)

type Receiver interface {
	Output(i GetInfo, level Level, fields map[string]interface{}, msg string)
}

type receiver struct {
	logger *stdlog.Logger
}

func (r *receiver) Output(i GetInfo, level Level, fields map[string]interface{},
	msg string) {
	var sb strings.Builder
	sb.WriteString(level.String())
	sb.WriteString(defaultSeparator)
	sb.WriteString(i.Prefix())
	if len(fields) > 0 {
		fs, err := json.Marshal(fields)
		if err == nil {
			sb.WriteString(defaultSeparator)
			sb.Write(fs)
		}
	}
	sb.WriteString(defaultSeparator)
	sb.WriteString(msg)
	r.logger.Println(sb.String())
}

func newDefaultReceiver() Receiver {
	ol := stdlog.New(os.Stderr, "", stdlog.LstdFlags)
	return &receiver{
		logger: ol,
	}
}
