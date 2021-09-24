package log

import (
	"encoding/json"
	stdlog "log"
	"os"
	"strings"
)

type Receiver interface {
	Output(entry *BasicEntry, lvl Level, msg string)
}

type receiver struct {
	logger *stdlog.Logger
}

func (r *receiver) Output(entry *BasicEntry, lvl Level, msg string) {
	var sb strings.Builder
	sb.WriteString(lvl.String())
	sb.WriteString(defaultSeparator)
	sb.WriteString(entry.Logger.Prefix)
	if len(entry.Fields) > 0 {
		fs, err := json.Marshal(entry.Fields)
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
