// Package remote implements a Dispatcher for remote (peer) dispatching.
// Subproblems that are dispatched are done so over an RPC.
package remote

import (
	"context"

	"github.com/openfga/openfga/internal/dispatch"
)

var _ dispatch.Dispatcher = (*remoteDispatcher)(nil)

type remoteDispatcher struct {

	// dispatchClient *protobuf.DispatcherService
}

// DispatchCheck implements dispatch.Dispatcher.
func (r *remoteDispatcher) DispatchCheck(
	ctx context.Context,
	req *dispatch.DispatchCheckRequest,
) (*dispatch.DispatchCheckResponse, error) {

	// use consistent hashing to determine where the request should go

	// dispatch the request over the 'dispatchClient'

	panic("unimplemented")
}

// DispatchReverseExpand implements dispatch.Dispatcher.
func (r *remoteDispatcher) DispatchReverseExpand(
	req *dispatch.DispatchReverseExpandRequest,
	stream dispatch.ReverseExpanderStream,
) error {

	// use consistent hashing to determine where the request should go

	// dispatch the request over the 'dispatchClient'

	panic("unimplemented")
}
