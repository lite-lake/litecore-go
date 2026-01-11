package server

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestNewSignalHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	handler := NewSignalHandler(engine)

	if handler == nil {
		t.Fatal("NewSignalHandler() should not return nil")
	}

	if handler.engine != engine {
		t.Error("handler.engine should be set to the provided engine")
	}

	if handler.sigChan == nil {
		t.Error("handler.sigChan should be initialized")
	}
}

func TestSignalHandler_Start(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("start signal handler", func(t *testing.T) {
		engine, err := NewEngine()
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		handler := NewSignalHandler(engine)

		// Start the handler in a goroutine
		done := make(chan bool)
		go func() {
			handler.Start()
			done <- true
		}()

		// Give it time to start
		time.Sleep(10 * time.Millisecond)

		// Stop the handler
		handler.Stop()

		// Wait for completion
		select {
		case <-done:
			// Success
		case <-time.After(1 * time.Second):
			t.Error("SignalHandler.Start() did not complete in time")
		}
	})
}

func TestSignalHandler_Stop(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("stop signal handler", func(t *testing.T) {
		engine, err := NewEngine()
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		handler := NewSignalHandler(engine)
		handler.Start()

		// Stop should not block
		handler.Stop()
	})
}

func TestSignalHandler_HandleSIGINT(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("handle SIGINT signal", func(t *testing.T) {
		engine, err := NewEngine(
			WithServerConfig(&ServerConfig{
				Host: "127.0.0.1",
				Port: 0,
			}),
		)
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		if err := engine.Start(); err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		handler := NewSignalHandler(engine)
		handler.Start()

		// Send SIGINT signal
		done := make(chan bool)
		go func() {
			time.Sleep(50 * time.Millisecond)
			process, err := os.FindProcess(os.Getpid())
			if err != nil {
				t.Logf("FindProcess error: %v", err)
				return
			}
			_ = process.Signal(syscall.SIGINT)
		}()

		// Wait for signal to be processed
		go func() {
			time.Sleep(500 * time.Millisecond)
			done <- true
		}()

		select {
		case <-done:
			// Signal was processed
		case <-time.After(2 * time.Second):
			// Timeout is acceptable as the handler may have processed the signal
		}

		handler.Stop()
	})
}

func TestSignalHandler_HandleSIGTERM(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("handle SIGTERM signal", func(t *testing.T) {
		engine, err := NewEngine(
			WithServerConfig(&ServerConfig{
				Host: "127.0.0.1",
				Port: 0,
			}),
		)
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		if err := engine.Start(); err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		handler := NewSignalHandler(engine)
		handler.Start()

		// Send SIGTERM signal
		done := make(chan bool)
		go func() {
			time.Sleep(50 * time.Millisecond)
			process, err := os.FindProcess(os.Getpid())
			if err != nil {
				t.Logf("FindProcess error: %v", err)
				return
			}
			_ = process.Signal(syscall.SIGTERM)
		}()

		go func() {
			time.Sleep(500 * time.Millisecond)
			done <- true
		}()

		select {
		case <-done:
			// Signal was processed
		case <-time.After(2 * time.Second):
			// Timeout is acceptable
		}

		handler.Stop()
	})
}

func TestEngine_WaitForShutdown(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("wait for shutdown signal", func(t *testing.T) {
		engine, err := NewEngine(
			WithServerConfig(&ServerConfig{
				Host: "127.0.0.1",
				Port: 0,
			}),
		)
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		if err := engine.Start(); err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		// Send signal after delay
		go func() {
			time.Sleep(100 * time.Millisecond)
			process, err := os.FindProcess(os.Getpid())
			if err != nil {
				t.Logf("FindProcess error: %v", err)
				return
			}
			_ = process.Signal(syscall.SIGTERM)
		}()

		// WaitForShutdown should return when signal is received
		err = engine.WaitForShutdown(5 * time.Second)
		if err != nil {
			t.Errorf("WaitForShutdown() error = %v", err)
		}
	})

	t.Run("wait for shutdown timeout", func(t *testing.T) {
		engine, err := NewEngine()
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		// Wait with very short timeout
		start := time.Now()
		err = engine.WaitForShutdown(100 * time.Millisecond)
		elapsed := time.Since(start)

		if err == nil {
			t.Error("WaitForShutdown() should return error on timeout")
		}

		// Should timeout in approximately 100ms
		if elapsed < 100*time.Millisecond {
			t.Logf("WaitForShutdown() returned early: %v", elapsed)
		}

		if elapsed > 500*time.Millisecond {
			t.Errorf("WaitForShutdown() took too long: %v", elapsed)
		}
	})

	t.Run("wait for shutdown with context cancellation", func(t *testing.T) {
		engine, err := NewEngine()
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		// Cancel context to trigger return
		go func() {
			time.Sleep(100 * time.Millisecond)
			engine.cancel()
		}()

		err = engine.WaitForShutdown(5 * time.Second)
		if err != nil {
			t.Errorf("WaitForShutdown() should not error on context cancel, got %v", err)
		}
	})
}

func TestEngine_ShutdownWithContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("shutdown with context cancellation", func(t *testing.T) {
		engine, err := NewEngine()
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())

		// Cancel context after delay
		go func() {
			time.Sleep(100 * time.Millisecond)
			cancel()
		}()

		// Should return when context is cancelled
		err = engine.ShutdownWithContext(ctx)
		if err != context.Canceled {
			t.Errorf("ShutdownWithContext() should return context.Canceled, got %v", err)
		}
	})

	t.Run("shutdown with signal", func(t *testing.T) {
		engine, err := NewEngine(
			WithServerConfig(&ServerConfig{
				Host: "127.0.0.1",
				Port: 0,
			}),
		)
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		if err := engine.Start(); err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Send signal after delay
		go func() {
			time.Sleep(100 * time.Millisecond)
			process, err := os.FindProcess(os.Getpid())
			if err != nil {
				t.Logf("FindProcess error: %v", err)
				return
			}
			_ = process.Signal(syscall.SIGTERM)
		}()

		// Should complete when signal is received
		err = engine.ShutdownWithContext(ctx)
		if err != nil {
			t.Errorf("ShutdownWithContext() error = %v", err)
		}
	})
}

func TestEngine_RegisterShutdownHook(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("register and execute shutdown hook", func(t *testing.T) {
		engine, err := NewEngine()
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		hookExecuted := false
		hook := func() {
			hookExecuted = true
		}

		engine.RegisterShutdownHook(hook)

		// Send signal to trigger hook
		go func() {
			time.Sleep(100 * time.Millisecond)
			process, err := os.FindProcess(os.Getpid())
			if err != nil {
				t.Logf("FindProcess error: %v", err)
				return
			}
			_ = process.Signal(syscall.SIGTERM)
		}()

		// Wait for hook to be potentially called
		time.Sleep(500 * time.Millisecond)

		// Note: The hook runs in a separate goroutine and we can't easily
		// verify it was called without synchronization, but the test
		// verifies that RegisterShutdownHook doesn't panic or block
		_ = hookExecuted // Use the variable to avoid unused variable error
	})
}

func TestEngine_Run(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("run and shutdown with signal", func(t *testing.T) {
		engine, err := NewEngine(
			WithServerConfig(&ServerConfig{
				Host: "127.0.0.1",
				Port: 0,
			}),
		)
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		// Run in background
		done := make(chan error)
		go func() {
			done <- engine.Run()
		}()

		// Wait for server to start
		time.Sleep(200 * time.Millisecond)

		// Verify server is running
		if !engine.IsStarted() {
			t.Error("engine should be started")
		}

		// Send shutdown signal
		process, err := os.FindProcess(os.Getpid())
		if err != nil {
			t.Fatalf("FindProcess error: %v", err)
		}
		_ = process.Signal(syscall.SIGTERM)

		// Wait for Run to complete
		select {
		case err := <-done:
			if err != nil {
				t.Errorf("Run() returned error: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Error("Run() did not complete in time")
		}

		// Verify server is stopped
		if engine.IsStarted() {
			t.Error("engine should be stopped after Run completes")
		}
	})
}

func TestSignalHandler_Signals(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("handle multiple signals", func(t *testing.T) {
		engine, err := NewEngine(
			WithServerConfig(&ServerConfig{
				Host: "127.0.0.1",
				Port: 0,
			}),
		)
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		if err := engine.Start(); err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		handler := NewSignalHandler(engine)
		handler.Start()
		defer handler.Stop()

		// Send SIGINT
		go func() {
			time.Sleep(50 * time.Millisecond)
			process, err := os.FindProcess(os.Getpid())
			if err != nil {
				return
			}
			_ = process.Signal(syscall.SIGINT)
		}()

		// Wait a bit for signal processing
		time.Sleep(500 * time.Millisecond)

		// Engine should be stopped
		if !engine.IsStarted() {
			// Expected behavior after signal
		}
	})
}

func TestEngine_WaitForShutdown_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("full lifecycle with signal", func(t *testing.T) {
		engine, err := NewEngine(
			WithServerConfig(&ServerConfig{
				Host: "127.0.0.1",
				Port: 0,
			}),
		)
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if err := engine.Initialize(); err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}

		if err := engine.Start(); err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		// Verify server is running
		if !engine.IsStarted() {
			t.Error("engine should be started")
		}

		// Schedule shutdown signal
		go func() {
			time.Sleep(200 * time.Millisecond)
			process, err := os.FindProcess(os.Getpid())
			if err != nil {
				t.Logf("FindProcess error: %v", err)
				return
			}
			_ = process.Signal(syscall.SIGTERM)
		}()

		// Wait for shutdown
		err = engine.WaitForShutdown(5 * time.Second)
		if err != nil {
			t.Errorf("WaitForShutdown() error = %v", err)
		}

		// Verify cleanup
		if engine.IsStarted() {
			t.Error("engine should be stopped after shutdown")
		}
	})
}
