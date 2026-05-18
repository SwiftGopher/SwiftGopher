package tests

import (
	"context"
	"sync"
	"testing"
	"time"

	"swift-gopher/internal/worker"
)

func TestPool_ExecutesJobs(t *testing.T) {
	p := worker.NewPool(3, 10, newTestLogger())
	p.Start(context.Background())

	var mu sync.Mutex
	sum := 0

	for i := 0; i < 5; i++ {
		v := i
		err := p.Submit(func() {
			mu.Lock()
			sum += v
			mu.Unlock()
		})
		if err != nil {
			t.Fatalf("submit error: %v", err)
		}
	}

	p.Shutdown(context.Background())

	if sum != 10 {
		t.Fatalf("expected 10 got %d", sum)
	}
}
func TestPool_QueueFull(t *testing.T) {
	p := worker.NewPool(1, 1, newTestLogger())
	p.Start(context.Background())

	block := make(chan struct{})

	_ = p.Submit(func() {
		<-block
	})

	_ = p.Submit(func() {})

	err := p.Submit(func() {})
	if err == nil {
		t.Fatalf("expected queue full error")
	}

	close(block)
	p.Shutdown(context.Background())
}
func TestPool_ShutdownWaits(t *testing.T) {
	p := worker.NewPool(2, 10, newTestLogger())
	p.Start(context.Background())

	err := p.Submit(func() {
		time.Sleep(100 * time.Millisecond)
	})
	if err != nil {
		t.Fatalf("submit error: %v", err)
	}

	done := make(chan struct{})

	go func() {
		p.Shutdown(context.Background())
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("shutdown timeout")
	}
}
func TestPool_PanicRecovery(t *testing.T) {
	p := worker.NewPool(1, 10, newTestLogger())
	p.Start(context.Background())

	_ = p.Submit(func() {
		panic("boom")
	})

	p.Shutdown(context.Background())
}
func TestPool_RejectAfterShutdown(t *testing.T) {
	p := worker.NewPool(1, 10, newTestLogger())
	p.Start(context.Background())

	p.Shutdown(context.Background())

	err := p.Submit(func() {})
	if err == nil {
		t.Fatalf("expected error after shutdown")
	}
}
