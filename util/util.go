package util

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func WaitSignal() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, os.Interrupt)
	DebugLog(fmt.Sprintf("terminate signal(%d) received", <-quit), 0)
	close(quit)
}
