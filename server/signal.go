package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// SignalHandler 信号处理器
type SignalHandler struct {
	engine  *Engine
	sigChan chan os.Signal
}

// NewSignalHandler 创建信号处理器
func NewSignalHandler(engine *Engine) *SignalHandler {
	return &SignalHandler{
		engine:  engine,
		sigChan: make(chan os.Signal, 1),
	}
}

// Start 开始监听信号
func (h *SignalHandler) Start() {
	signal.Notify(h.sigChan,
		syscall.SIGINT,  // Ctrl+C
		syscall.SIGTERM, // kill
		syscall.SIGQUIT, // quit
		syscall.SIGHUP,  // hangup
	)

	go h.handleSignals()
}

// handleSignals 处理信号
func (h *SignalHandler) handleSignals() {
	for sig := range h.sigChan {
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			fmt.Printf("Received signal %v, shutting down...\n", sig)
			if err := h.engine.Stop(); err != nil {
				fmt.Printf("Shutdown error: %v\n", err)
				os.Exit(1)
			}
			return

		case syscall.SIGQUIT:
			fmt.Printf("Received signal %v, shutting down...\n", sig)
			if err := h.engine.Stop(); err != nil {
				fmt.Printf("Shutdown error: %v\n", err)
				os.Exit(1)
			}
			return

		case syscall.SIGHUP:
			// 重载配置（可选实现）
			fmt.Println("Received SIGHUP, reloading configuration...")
			// 这里可以添加配置重载逻辑

		default:
			fmt.Printf("Received unhandled signal: %v\n", sig)
		}
	}
}

// Stop 停止信号处理器
func (h *SignalHandler) Stop() {
	signal.Stop(h.sigChan)
	close(h.sigChan)
}

// WaitForShutdown 等待关闭信号（带超时）
func (e *Engine) WaitForShutdown(timeout time.Duration) error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case sig := <-sigChan:
		fmt.Printf("Received signal %v, shutting down...\n", sig)
		return e.Stop()
	case <-time.After(timeout):
		return fmt.Errorf("timeout waiting for shutdown signal")
	case <-e.ctx.Done():
		return nil
	}
}

// ShutdownWithContext 使用上下文进行优雅关闭
func (e *Engine) ShutdownWithContext(ctx context.Context) error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case sig := <-sigChan:
		fmt.Printf("Received signal %v, shutting down...\n", sig)
		return e.Stop()
	case <-ctx.Done():
		return ctx.Err()
	}
}

// RegisterShutdownHook 注册关闭钩子
func (e *Engine) RegisterShutdownHook(hook func()) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sigChan
		hook()
	}()
}
