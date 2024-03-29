package dispatch

import "context"

type CheckDispatcher interface {
	DispatchCheck(
		ctx context.Context,
		req *DispatchCheckRequest,
	) (*DispatchCheckResponse, error)
}

type DispatchCheckRequest struct {
	// other request fields here ...

	Metadata DispatchCheckRequestMetadata
}

func (d *DispatchCheckRequest) GetDispatchCount() uint32 {
	return d.Metadata.DispatchCount
}

type DispatchCheckRequestMetadata struct {
	DispatchCount uint32
}

type DispatchCheckResponse struct {
	// response fields here ...
}
