package gemini

import (
	"context"
	"math/rand/v2"
	"time"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	retryBaseDelay  = 5 * time.Second
	retryMaxDelay   = 90 * time.Second // covers the ~53s API-suggested delay for quota resets
	retryMultiplier = 2.0
	retryJitter     = 0.25 // ±25% of the computed delay
)

// withRetry calls fn up to maxAttempts times, retrying on transient Gemini API
// errors with exponential backoff and jitter. When the API provides an explicit
// retry delay via RetryInfo, that delay is used in preference to the computed one.
// Returns the last error if all attempts are exhausted, or ctx.Err() if cancelled.
func withRetry(ctx context.Context, maxAttempts int, fn func() error) error {
	var err error
	delay := retryBaseDelay

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}

		if !isRetryable(err) {
			return err
		}

		if attempt == maxAttempts {
			break
		}

		// Use the API-provided retry delay when available (e.g. 429 quota resets),
		// otherwise fall back to exponential backoff with jitter.
		sleep := retryDelayFromError(err)
		if sleep <= 0 {
			jitter := 1.0 + retryJitter*(2*rand.Float64()-1)
			sleep = time.Duration(float64(delay) * jitter)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(sleep):
		}

		delay = time.Duration(float64(delay) * retryMultiplier)
		if delay > retryMaxDelay {
			delay = retryMaxDelay
		}
	}

	return err
}

// retryDelayFromError extracts the server-recommended retry delay from a gRPC
// status error containing a RetryInfo detail (as returned by Gemini on 429s).
// Returns 0 if no RetryInfo is present or the error is not a gRPC status error.
func retryDelayFromError(err error) time.Duration {
	st, ok := status.FromError(err)
	if !ok {
		return 0
	}
	for _, detail := range st.Details() {
		if ri, ok := detail.(*errdetails.RetryInfo); ok {
			if d := ri.GetRetryDelay(); d != nil {
				return d.AsDuration()
			}
		}
	}
	return 0
}

// isRetryable reports whether an error from the Gemini API is transient and
// worth retrying. Permanent errors (invalid argument, not found, etc.) are not
// retried because repeating the request would produce the same outcome.
// Non-gRPC errors (e.g. JSON parse failures from truncated responses) are also
// considered retryable since they indicate a degraded API response.
func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	st, ok := status.FromError(err)
	if !ok {
		// Not a gRPC status error — treat as retryable (e.g. JSON parse of truncated response).
		return true
	}
	switch st.Code() {
	case codes.Unavailable,      // 503 — service temporarily unavailable / high demand
		codes.ResourceExhausted, // 429 — token quota exceeded; backoff and retry
		codes.DeadlineExceeded,  // timeout on the server side
		codes.Aborted:           // transient conflict; safe to retry
		return true
	default:
		return false
	}
}

