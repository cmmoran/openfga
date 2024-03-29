package cached

import (
	"context"

	"github.com/openfga/openfga/internal/dispatch"
)

var _ dispatch.Dispatcher = (*cachedDispatcher)(nil)

type cachedDispatcher struct {
	delegate dispatch.Dispatcher

	checkCache cache[string, *dispatch.DispatchCheckResponse]
}

type cache[K string, V any] interface {
	Set(key K, value V)
	Get(key K) (V, bool)
}

func NewCachedDispatcher() *cachedDispatcher {
	return &cachedDispatcher{}
}

func (c *cachedDispatcher) SetDelegate(delegate dispatch.Dispatcher) {
	c.delegate = delegate
}

// DispatchCheck implements dispatch.Dispatcher.
func (c *cachedDispatcher) DispatchCheck(
	ctx context.Context,
	req *dispatch.DispatchCheckRequest,
) (*dispatch.DispatchCheckResponse, error) {
	var key string // compute cache key given 'req'

	if val, ok := c.checkCache.Get(key); ok {
		return val, nil
	}

	return c.delegate.DispatchCheck(ctx, req)
}

// DispatchReverseExpand implements dispatch.Dispatcher.
func (c *cachedDispatcher) DispatchReverseExpand(
	req *dispatch.DispatchReverseExpandRequest,
	stream dispatch.ReverseExpanderStream,
) error {
	// no caching for ReverseExpand today, we'd have to think about that more
	return c.delegate.DispatchReverseExpand(req, stream)
}
