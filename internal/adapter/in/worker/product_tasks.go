package worker

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/hibiken/asynq"
	"github.com/nuzirwan/go-boilerplate/internal/domain/port"
)

const (
	TaskApplyDiscount = "product:apply_discount"
)

type ProductTaskHandler struct {
	applyDiscount port.ApplyDiscount
}

func NewProductTaskHandler(applyDiscount port.ApplyDiscount) *ProductTaskHandler {
	return &ProductTaskHandler{applyDiscount: applyDiscount}
}

func (h *ProductTaskHandler) HandleApplyDiscount(ctx context.Context, t *asynq.Task) error {
	var cmd port.ApplyDiscountCommand
	if err := sonic.Unmarshal(t.Payload(), &cmd); err != nil {
		return fmt.Errorf("unmarshal task payload: %w", err)
	}
	_, err := h.applyDiscount.Execute(ctx, cmd)
	if err != nil {
		return fmt.Errorf("apply discount: %w", err)
	}
	return nil
}

func (h *ProductTaskHandler) Register(mux *asynq.ServeMux) {
	mux.HandleFunc(TaskApplyDiscount, h.HandleApplyDiscount)
}
