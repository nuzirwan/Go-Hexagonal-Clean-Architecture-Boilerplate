package observability

import (
	"context"
	"errors"
	"testing"

	"go.opentelemetry.io/otel/attribute"
)

func TestStartOp_Success(t *testing.T) {
	ctx := context.Background()

	ctx, end := StartOp(ctx, "usecase.Test", attribute.String("key", "value"))
	if ctx == nil {
		t.Fatal("expected non-nil context")
	}

	// Should not panic on success path
	end(nil, attribute.String("result", "ok"))
}

func TestStartOp_Error(t *testing.T) {
	ctx := context.Background()

	ctx, end := StartOp(ctx, "usecase.TestError", attribute.String("id", "123"))
	if ctx == nil {
		t.Fatal("expected non-nil context")
	}

	// Should not panic on error path
	end(errors.New("something went wrong"))
}

func TestStartOp_NoAttrs(t *testing.T) {
	ctx := context.Background()

	ctx, end := StartOp(ctx, "usecase.TestNoAttrs")
	if ctx == nil {
		t.Fatal("expected non-nil context")
	}

	end(nil)
}
