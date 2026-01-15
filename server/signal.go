package server

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// WaitForShutdown 等待关闭信号
func (e *Engine) WaitForShutdown() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	sig := <-sigs
	fmt.Printf("Received signal %v, shutting down...\n", sig)

	if err := e.Stop(); err != nil {
		fmt.Printf("shutdown error: %v\n", err)
		os.Exit(1)
	}
}
