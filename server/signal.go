package server

import (
	"os"
	"os/signal"
	"syscall"
)

// WaitForShutdown 等待关闭信号
func (e *Engine) WaitForShutdown() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	sig := <-sigs
	e.builtin.LoggerManager.Logger("server").Info("Received shutdown signal", "signal", sig)

	if err := e.Stop(); err != nil {
		e.builtin.LoggerManager.Logger("server").Fatal("Shutdown error", "error", err)
		os.Exit(1)
	}
}
