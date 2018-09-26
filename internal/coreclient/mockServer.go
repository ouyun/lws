package coreclient

import "io"

type MockServer struct {
	WireReq wireRequest
	WireRes wireResponse
	Encoder *messageEncoder
	Decoder *messageDecoder
}

func NewMockServer(conn io.ReadWriteCloser) *MockServer {
	this := &MockServer{
		WireReq: wireRequest{},
		WireRes: wireResponse{},
		Encoder: newMessageEncoder(conn, 1024),
		Decoder: newMessageDecoder(conn, 1024),
	}
	return this
}
