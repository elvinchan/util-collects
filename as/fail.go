package as

import (
	"fmt"
	"go/ast"
	"reflect"
	"runtime"
	"strings"
)

func (c *As) fail(title string, details []string, msgAndArgs ...interface{}) {
	c.Helper()

	// write info
	c.info.WriteByte('\n')
	c.info.WriteString("error:\n")
	c.writeError(title, details)
	c.info.WriteString("message:\n")
	c.writeMessage(msg(msgAndArgs...))
	c.info.WriteString("stack:\n")
	c.writeStack()
	c.TB.Error(c.info.String())
	c.info.Reset()

	if c.failDirectly {
		c.FailNow()
	} else {
		c.Fail()
	}
}

func msg(msgAndArgs ...interface{}) (v string) {
	if len(msgAndArgs) == 1 {
		msg := msgAndArgs[0]
		if msgAsStr, ok := msg.(string); ok {
			v = msgAsStr
		} else {
			v = fmt.Sprintf("%+v", msg)
		}
	}
	if len(msgAndArgs) > 1 {
		v = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
	}
	return
}

func (c *As) writeError(title string, details []string) {
	fmt.Fprintf(&c.info, "%s%s\n", prefix, title)
	for _, detail := range details {
		fmt.Fprintf(&c.info, "%s%s%s\n", prefix, prefix, detail)
	}
}

func (c *As) writeMessage(message string) {
	fmt.Fprintf(&c.info, "%s%s\n", prefix, message)
}

func (c *As) writeStack() {
	pc := make([]uintptr, 8)
	runtime.Callers(c.skipCallers, pc)
	frames := runtime.CallersFrames(pc)
	thisPackage := reflect.TypeOf(As{}).PkgPath() + "."
	for {
		frame, more := frames.Next()
		if strings.HasPrefix(frame.Function, "testing.") {
			// Stop before getting back to stdlib test runner calls.
			break
		}
		if fname := strings.TrimPrefix(
			frame.Function, thisPackage,
		); fname != frame.Function {
			if ast.IsExported(fname) {
				// Continue without printing frames for as exported API.
				continue
			}
			// Stop when entering as internal calls.
			break
		}
		fmt.Fprintf(&c.info, "%s%s:%d\n", prefix, frame.File, frame.Line)
		if !more {
			// There are no more callers.
			break
		}
	}
}

// prefix is the string used to indent blocks of output.
const prefix = "  "
