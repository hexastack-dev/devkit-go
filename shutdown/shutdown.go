package shutdown

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/hexastack-dev/devkit-go/log"
)

type signalNotifier interface {
	Notify(c chan<- os.Signal, sig ...os.Signal)
}

type osSignalNotifierFunc func(c chan<- os.Signal, sig ...os.Signal)

func (n osSignalNotifierFunc) Notify(c chan<- os.Signal, sig ...os.Signal) {
	n(c, sig...)
}

type Shutdown struct {
	timeout   time.Duration
	listeners map[string]Listener
	notifier  signalNotifier
}

type Listener interface {
	OnShutdown(context.Context) error
}

type ListenerFunc func(context.Context) error

func (f ListenerFunc) OnShutdown(ctx context.Context) error {
	return f(ctx)
}

func New(timeout time.Duration, listeners map[string]Listener) *Shutdown {
	return &Shutdown{
		timeout:   timeout,
		listeners: listeners,
		notifier:  osSignalNotifierFunc(signal.Notify),
	}
}

// Wait receive an message that will be printed out to logger before blocking.
// Wait block the call and wait for SIGINT, SIGTERM or SIGHUP signals.
// After signal is received, this method will run Listener.OnShutdown in parallel,
// the operation must complete under specified timeout. If any listeners return an
// error or operation is timed out this method will call os.Exit(1).
func (s *Shutdown) Wait() (os.Signal, error) {
	quit := make(chan os.Signal, 1)
	defer close(quit)

	s.notifier.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	// wait for shutdown signals
	sig := <-quit
	log.Debug(fmt.Sprintf("Received signal %s, shutting down", sig.String()))
	// timeoutFunc := time.AfterFunc(s.timeout, func() {
	// 	err := fmt.Errorf("shutdown did not complete after %d%s", s.timeout.Milliseconds(), "ms")
	// 	getLogger(s.logger).Error("Shutdown timeout", err)
	// 	os.Exit(1)
	// })
	// defer timeoutFunc.Stop()

	done := make(chan bool, 1)

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	go func() {
		done <- s.onShutdown(ctx)
	}()

	select {
	case <-ctx.Done():
		err := fmt.Errorf("shutdown did not complete after %d%s: %w", s.timeout.Milliseconds(), "ms", ctx.Err())
		return sig, err
	case ok := <-done:
		if !ok {
			return sig, errors.New("shutdown completed with error")
		} else {
			return sig, nil
		}
	}
}

func (s *Shutdown) onShutdown(ctx context.Context) bool {
	var (
		wg sync.WaitGroup
		ok bool
	)
	for name, listener := range s.listeners {
		wg.Add(1)
		go func(name string, listener Listener) {
			defer wg.Done()

			log.Debug(fmt.Sprintf("Running shutdown listener: %s", name))
			if err := listener.OnShutdown(ctx); err != nil {
				log.Error(fmt.Sprintf("Shutdown listener %s return an error", name), err)
			} else {
				ok = true
			}
		}(name, listener)
	}
	wg.Wait()
	return ok
}
