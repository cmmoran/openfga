package server

import (
	localdispatch "github.com/openfga/openfga/internal/dispatch/local"
)

func NewServer() *Server {
	localDispatcher := localdispatch.NewLocalDispatcher()

	s := Server{
		dispatcher: localDispatcher,
	}

	return &s
}
