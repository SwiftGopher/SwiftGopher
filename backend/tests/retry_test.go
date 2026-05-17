package tests

import (
	"errors"
	"testing"
	"time"
)

func retry(attempts int, delay time.Duration, fn func() error) error {
	for i := 0; i < attempts; i++ {
		if err := fn(); err == nil {
			return nil
		}
		if i < attempts-1 {
			time.Sleep(delay)
		}
	}
	return errors.New("max attempts reached")
}

func TestRetry_SuccessOnFirstAttempt(t *testing.T) {
	callCount := 0
	err := retry(3, 10*time.Millisecond, func() error {
		callCount++
		return nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}
}

func TestRetry_SuccessOnSecondAttempt(t *testing.T) {
	callCount := 0
	err := retry(3, 10*time.Millisecond, func() error {
		callCount++
		if callCount == 1 {
			return errors.New("temporary error")
		}
		return nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls, got %d", callCount)
	}
}

func TestRetry_FailureAfterMaxAttempts(t *testing.T) {
	callCount := 0
	err := retry(2, 10*time.Millisecond, func() error {
		callCount++
		return errors.New("persistent error")
	})

	if err == nil {
		t.Error("expected error, got nil")
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls, got %d", callCount)
	}
}
