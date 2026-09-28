package network

import (
	"context"
	"errors"
	"math/rand"
	"net/http"
	"time"
)

// Backoff computes exponential backoff with jitter.
func Backoff(attempt int) time.Duration {
	base := 500 * time.Millisecond
	if attempt < 0 {
		attempt = 0
	}
	if attempt > 6 {
		attempt = 6
	}

	delay := base * time.Duration(1<<attempt)
	const maxDelay = 8 * time.Second
	if delay > maxDelay {
		delay = maxDelay
	}

	// Add 10-20% jitter to prevent thundering herd
	jitter := time.Duration(rand.Int63n(int64(delay / 5)))
	return delay + jitter
}

// IsRetryable determines if a network error represents a temporary/transient failure.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	var netErr *NetworkError
	if errors.As(err, &netErr) {
		if netErr.Kind == ErrorCancelled {
			return false
		}
		if netErr.Kind == ErrorTimeout || netErr.Kind == ErrorNetwork {
			return true
		}
		if netErr.Kind == ErrorHTTP {
			switch netErr.StatusCode {
			case http.StatusRequestTimeout,
				http.StatusTooManyRequests,
				http.StatusInternalServerError,
				http.StatusBadGateway,
				http.StatusServiceUnavailable,
				http.StatusGatewayTimeout:
				return true
			default:
				// 400, 401, 403, 404, 405, etc. are NOT retryable
				return false
			}
		}
	}

	return false
}

// Retry executes the given operation up to maxAttempts times with exponential backoff and context cancellation.
func Retry(ctx context.Context, maxAttempts int, op func(ctx context.Context, attempt int) error) error {
	if maxAttempts <= 0 {
		maxAttempts = 1
	}

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err := op(ctx, attempt)
		if err == nil {
			return nil
		}
		lastErr = err

		if !IsRetryable(err) || attempt == maxAttempts-1 {
			return err
		}

		backoffDuration := Backoff(attempt)
		timer := time.NewTimer(backoffDuration)

		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}

	return lastErr
}
