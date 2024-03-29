// Package cycle implements a Dispatcher with cycle detection wrapping
// the dispatches.
package cycle

import (
	"context"
	"fmt"

	"github.com/openfga/openfga/internal/dispatch"
)

type cycleDetectionDispatcher struct {
	delegate dispatch.Dispatcher
}

func NewCycleDetectectionDispatcher() *cycleDetectionDispatcher {
	return &cycleDetectionDispatcher{}
}

func (c *cycleDetectionDispatcher) SetDelegate(delegate dispatch.Dispatcher) {
	c.delegate = delegate
}

// DispatchCheck implements dispatch.Dispatcher.
func (c *cycleDetectionDispatcher) DispatchCheck(
	ctx context.Context,
	req *dispatch.DispatchCheckRequest,
) (*dispatch.DispatchCheckResponse, error) {
	var cycle bool // logic to determine cycle here

	if cycle {
		return nil, fmt.Errorf("cycle detected")
	}

	return c.delegate.DispatchCheck(ctx, req)
}

// DispatchReverseExpand implements dispatch.Dispatcher.
func (c *cycleDetectionDispatcher) DispatchReverseExpand(
	req *dispatch.DispatchReverseExpandRequest,
	stream dispatch.ReverseExpanderStream,
) error {
	var cycle bool // logic to determine cycle here

	if cycle {
		return fmt.Errorf("cycle detected")
	}

	return c.delegate.DispatchReverseExpand(req, stream)
}
