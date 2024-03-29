package throttled

import (
	"context"

	"github.com/openfga/openfga/internal/dispatch"
)

var _ dispatch.Dispatcher = (*throttledDispatcher)(nil)

type throttledDispatcher struct {
	delegate dispatch.Dispatcher

	dispatchThreshold uint32

	checkThrottler         dispatchThrottler[*dispatch.DispatchCheckRequest]
	reverseExpandThrottler dispatchThrottler[*dispatch.DispatchReverseExpandRequest]
}

func NewThrottledDispatcher() *throttledDispatcher {
	return &throttledDispatcher{}
}

func (t *throttledDispatcher) SetDelegate(delegate dispatch.Dispatcher) {
	t.delegate = delegate
}

// DispatchCheck implements dispatch.Dispatcher.
func (t *throttledDispatcher) DispatchCheck(
	ctx context.Context,
	req *dispatch.DispatchCheckRequest,
) (*dispatch.DispatchCheckResponse, error) {
	if req.GetDispatchCount() >= t.dispatchThreshold {
		// throttle, for example
		t.checkThrottler.delay()
	}

	return t.delegate.DispatchCheck(ctx, req)
}

// DispatchReverseExpand implements dispatch.Dispatcher.
func (t *throttledDispatcher) DispatchReverseExpand(
	req *dispatch.DispatchReverseExpandRequest,
	stream dispatch.ReverseExpanderStream,
) error {
	if req.GetDispatchCount() >= t.dispatchThreshold {
		// throttle, for example
		t.reverseExpandThrottler.delay()
	}

	return t.delegate.DispatchReverseExpand(req, stream)
}

type throttableRequest interface {
	GetDispatchCount() uint32
}

type dispatchThrottler[R throttableRequest] struct {
	throttleQueue chan R
}

// implement a common dispatch throttler given the generic type R
func (d *dispatchThrottler[R]) delay() {
	// delay, for example by waiting on some queue or something
	_ = d.throttleQueue
}
