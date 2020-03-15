package network

import (
	"io"
)

type PayloadInterface interface {
	Decode(r io.Reader, protocolVersion uint32) error
	Encode(w io.Writer, protocolVersion uint32) error
}

// //make sure version payload implement the interfae
// var _ PayloadInterface = (*payload.Version)(nil)

// func NewPayloadVersion() PayloadInterface {
// 	return &payload.Version{}
// }
