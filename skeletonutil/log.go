package skeletonutil

import (
	"fmt"
	"github.com/rs/zerolog"
	"os"
	"runtime"
)

var zerologStr = zerolog.New(os.Stderr).With().Timestamp().Logger()

func SetLogger(l zerolog.Logger) {
	zerologStr = l
}

func DebugLog(msg string, addStack int) {
	zerologStr.Debug().Msg(decolateLog(msg, addStack))
}

func InfoLog(msg string, addStack int) {
	zerologStr.Info().Msg(decolateLog(msg, addStack))
}

func ErrorLog(msg string, addStack int) {
	zerologStr.Error().Msg(decolateLog(msg, addStack))
}

func decolateLog(msg string, addStack int) string {
	_, file, line, ok := runtime.Caller(2 + addStack) //this, **Log, caller
	if !ok {
		return msg
	} else {
		return fmt.Sprintf("file: %s, line: %d, msg: %s", file, line, msg)
	}
}
