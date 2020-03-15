package network

import (
	"bytes"
	"fmt"
	"io"
	"time"
)

type Ping struct {
	Timestamp   uint32      //4 bytes
	Nonce       uint32      //4 bytes
	BlockHeight uint32      //4 bytes
}

func NewPingPayload(nonce uint32) *Ping {

	v := Ping{}
	v.Timestamp = uint32(time.Unix(time.Now().Unix(), 0).Unix())
	v.Nonce = nonce
	v.BlockHeight = 0 //TODO get this from a chain data
	return &v
}

func (v *Ping) Decode(r io.Reader, protocolVersion uint32) error {

	buf, ok := r.(*bytes.Buffer)

	if !ok {
		return fmt.Errorf("Ping decode reader is not a *byte.Buffer")
	}

	err := ReadElements(buf, &v.BlockHeight, (*uint32)(&v.Timestamp), &v.Nonce)
	if err != nil {
		return err
	}

	return nil
}

func (v *Ping) Encode(w io.Writer, protocolVersion uint32) error {
	err := WriteElement(w, v.BlockHeight)
	if err != nil {
		return err
	}

	err = WriteElement(w, v.Timestamp)
	if err != nil {
		return err
	}

	err = WriteElement(w, v.Nonce)
	if err != nil {
		return err
	}

	return nil
}
