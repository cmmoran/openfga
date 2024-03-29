package local

import (
	"context"

	"github.com/openfga/openfga/internal/dispatch"
	"github.com/openfga/openfga/internal/dispatch/cached"
	"github.com/openfga/openfga/internal/dispatch/cycle"
	"github.com/openfga/openfga/internal/dispatch/throttled"
	"github.com/openfga/openfga/internal/graph"
	"github.com/openfga/openfga/pkg/server/commands/reverseexpand"
)

var _ dispatch.Dispatcher = (*localDispatcher)(nil)

type localDispatcher struct {
	checker         *graph.LocalChecker
	reverseExpander *reverseexpand.ReverseExpandQuery

	checkQueryCacheEnabled    bool
	dispatchThrottlingEnabled bool
}

type LocalDispatcherOption func(*localDispatcher)

func WithCheckQueryCacheEnabled() LocalDispatcherOption {
	return func(ld *localDispatcher) {
		ld.checkQueryCacheEnabled = true
	}
}

func WithDispatchThrottlingEnabled() LocalDispatcherOption {
	return func(ld *localDispatcher) {
		ld.dispatchThrottlingEnabled = true
	}
}

// NewLocalDispatcher constructs a dispatch.Dispatcher which
// delegates locally through functiona call loopback for
// resolving dispatched subproblems.
func NewLocalDispatcher(opts ...LocalDispatcherOption) dispatch.Dispatcher {
	var checker *graph.LocalChecker

	localDispatcher := &localDispatcher{
		checker: checker,
	}

	for _, opt := range opts {
		opt(localDispatcher)
	}

	cycleDispatcher := cycle.NewCycleDetectectionDispatcher()
	throttledDispatcher := throttled.NewThrottledDispatcher()
	cachedDispatcher := cached.NewCachedDispatcher()

	cycleDispatcher.SetDelegate(localDispatcher)

	if localDispatcher.dispatchThrottlingEnabled {
		cycleDispatcher.SetDelegate(throttledDispatcher)
		throttledDispatcher.SetDelegate(localDispatcher)
	}

	if localDispatcher.checkQueryCacheEnabled {
		throttledDispatcher.SetDelegate(cachedDispatcher)
		cachedDispatcher.SetDelegate(localDispatcher)

		if !localDispatcher.dispatchThrottlingEnabled {
			cycleDispatcher.SetDelegate(cachedDispatcher)
		}
	}

	checker.SetDelegate(cycleDispatcher)

	return localDispatcher
}

// DispatchCheck implements dispatch.Dispatcher.
func (l *localDispatcher) DispatchCheck(
	ctx context.Context,
	req *dispatch.DispatchCheckRequest,
) (*dispatch.DispatchCheckResponse, error) {
	_ = l.checker // use the implementation of the checker to resolve the dispatch

	panic("not implemented")
}

// DispatchReverseExpand implements dispatch.Dispatcher.
func (l *localDispatcher) DispatchReverseExpand(
	req *dispatch.DispatchReverseExpandRequest,
	stream dispatch.ReverseExpanderStream,
) error {
	_ = l.reverseExpander // use the implementation of the reverse expander

	panic("not implemented")
}

func (l *localDispatcher) Close() {}
