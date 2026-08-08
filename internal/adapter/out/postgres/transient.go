package postgres

import (
	"database/sql"
	"errors"
	"strings"
)

// nonRetryable wraps errors that should not be retried.
// Use to short-circuit retry loops for domain/logic errors.
type nonRetryable struct{ err error }

func (e *nonRetryable) Error() string { return e.err.Error() }
func (e *nonRetryable) Unwrap() error { return e.err }

// isTransient returns true for connection-level errors that are safe to retry.
// NOT safe to retry: constraint violations, not found, syntax errors, deadlocks.
func isTransient(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, sql.ErrConnDone) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "bad connection") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "broken pipe")
}
