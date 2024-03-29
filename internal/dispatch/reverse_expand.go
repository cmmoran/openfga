package dispatch

type ReverseExpandDispatcher interface {
	DispatchReverseExpand(
		req *DispatchReverseExpandRequest,
		stream ReverseExpanderStream,
	) error
}

type ReverseExpanderStream interface {
	Send(*DispatchedReverseExpandResponse) error
}

type DispatchReverseExpandRequest struct {
	// other request fields here...

	Metadata DispatchReverseExpandRequestMetadata
}

func (d *DispatchReverseExpandRequest) GetDispatchCount() uint32 {
	return d.Metadata.DispatchCount
}

type DispatchReverseExpandRequestMetadata struct {
	DispatchCount uint32
}

type DispatchedReverseExpandResponse struct {
	// response fields here...
}
