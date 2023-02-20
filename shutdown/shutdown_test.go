package shutdown

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"
)

type signalConsumer struct {
	c   chan<- os.Signal
	sig []os.Signal
}

type signalNotifierTest struct {
	consumers []*signalConsumer
	mutex     sync.Mutex
}

func (s *signalNotifierTest) Notify(c chan<- os.Signal, sig ...os.Signal) {
	s.mutex.Lock()
	con := &signalConsumer{
		c:   c,
		sig: sig,
	}
	s.consumers = append(s.consumers, con)
	s.mutex.Unlock()
}

func (s *signalNotifierTest) notify(sig os.Signal) {
	for _, con := range s.consumers {
		for _, sig0 := range con.sig {
			if sig0 == sig {
				con.c <- sig
				break
			}
		}
	}
}

func TestShutdown(t *testing.T) {
	notifier := &signalNotifierTest{}
	res := make(chan string, 1)
	defer close(res)

	var (
		sleep = 10 * time.Millisecond
		wg    sync.WaitGroup
	)

	wg.Add(1)
	// positive test
	go func() {
		defer wg.Done()
		listeners := make(map[string]Listener)
		listeners["test"] = ListenerFunc(func(ctx context.Context) error {
			res <- "ok"
			return nil
		})
		sh := New(2*sleep, listeners)
		sh.notifier = notifier
		sig, err := sh.Wait()
		if err != nil {
			t.Errorf("should not return any error: %v", err)
		}
		if sig != os.Interrupt {
			t.Errorf("should received os.Interrupt: %v", sig.String())
		}
		if out := <-res; out != "ok" {
			t.Errorf("result should equals ok: %s", out)
		}
	}()

	wg.Add(1)
	// negative test, should be timeout
	go func() {
		defer wg.Done()

		listeners := make(map[string]Listener)
		listeners["test"] = ListenerFunc(func(ctx context.Context) error {
			time.Sleep(2 * sleep)
			return nil
		})

		sh := New(time.Millisecond, listeners)
		sh.notifier = notifier
		_, err := sh.Wait()
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("should return timeout error: context deadline exceeded: %v", err)
		}
	}()

	go func() {
		time.Sleep(sleep)
		notifier.notify(os.Interrupt)
	}()

	wg.Wait()
}
