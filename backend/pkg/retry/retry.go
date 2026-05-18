package retry

import (
	"context"
	"errors"
	"time"
)

type NonRetryableError struct {
	Err error
}

func (e *NonRetryableError) Error() string { return e.Err.Error() }
func (e *NonRetryableError) Unwrap() error { return e.Err }

func IsNonRetryable(err error) bool {
	var nre *NonRetryableError
	return errors.As(err, &nre)
}

func Do(ctx context.Context, maxAttempts int, baseDelay time.Duration, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		lastErr = fn()

		if lastErr == nil {
			return nil
		}

		if IsNonRetryable(lastErr) {
			return lastErr
		}

		if attempt == maxAttempts-1 {
			break
		}

		delay := baseDelay * (1 << uint(attempt))

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	return lastErr
}
