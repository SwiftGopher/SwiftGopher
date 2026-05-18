package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	rpkg "swift-gopher/pkg/retry"
)

func TestRetrySuccessOnFirstAttempt(t *testing.T) {
	callCount := 0

	err := rpkg.Do(
		context.Background(),
		3,
		10*time.Millisecond,
		func() error {
			callCount++
			return nil
		},
	)

	if err != nil {
		t.Errorf("expected no error got %v", err)
	}

	if callCount != 1 {
		t.Errorf("expected 1 call got %d", callCount)
	}
}

func TestRetrySuccessOnSecondAttempt(t *testing.T) {
	callCount := 0

	err := rpkg.Do(
		context.Background(),
		3,
		10*time.Millisecond,
		func() error {
			callCount++

			if callCount == 1 {
				return errors.New("temporary error")
			}

			return nil
		},
	)

	if err != nil {
		t.Errorf("expected no error got %v", err)
	}

	if callCount != 2 {
		t.Errorf("expected 2 calls got %d", callCount)
	}
}

func TestRetryFailureAfterMaxAttempts(t *testing.T) {
	callCount := 0

	err := rpkg.Do(
		context.Background(),
		2,
		10*time.Millisecond,
		func() error {
			callCount++
			return errors.New("persistent error")
		},
	)

	if err == nil {
		t.Error("expected error got nil")
	}

	if callCount != 2 {
		t.Errorf("expected 2 calls got %d", callCount)
	}
}

func TestRetryStopsOnNonRetryable(t *testing.T) {
	callCount := 0

	err := rpkg.Do(
		context.Background(),
		5,
		10*time.Millisecond,
		func() error {
			callCount++

			return &rpkg.NonRetryableError{
				Err: errors.New("fatal"),
			}
		},
	)

	if err == nil {
		t.Error("expected error got nil")
	}

	if callCount != 1 {
		t.Errorf("expected 1 call got %d", callCount)
	}
}
