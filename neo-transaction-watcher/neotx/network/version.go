package network

import (
	"bytes"
	"fmt"
	"io"
	"time"
)

type ServiceFlag uint64

const (
	MaxUserAgentLength = 1024
	DefaultUserAgent   = "/Neo:2.10.2/" //test
	ProtocolVersion    = uint32(0)
	DefaultServiceFlag = ServiceFlag(1)
)

type Version struct {
	Version     uint32      //4 bytes
	Services    ServiceFlag //8 bytes
	Timestamp   uint32      //4 bytes
	Port        uint16      //2 bytes
	Nonce       uint32      //4 bytes
	UserAgent   string      //? bytes
	StartHeight uint32      //4 bytes
	Relay       bool        //Whether to receive and forward
}

func NewVersionPayload(port uint16, nonce uint32) *Version {

	v := Version{}
	v.Version = ProtocolVersion
	v.Services = DefaultServiceFlag
	v.Timestamp = uint32(time.Unix(time.Now().Unix(), 0).Unix())
	v.Port = port
	v.Nonce = nonce
	v.UserAgent = DefaultUserAgent
	v.StartHeight = 0 //TODO get this from a chain data
	v.Relay = true
	return &v
}

func (v *Version) Decode(r io.Reader, protocolVersion uint32) error {

	buf, ok := r.(*bytes.Buffer)

	if !ok {
		return fmt.Errorf("Version decode reader is not a *byte.Buffer")
	}

	err := ReadElements(buf, &v.Version, &v.Services,
		(*uint32)(&v.Timestamp), &v.Port, &v.Nonce)
	if err != nil {
		return err
	}

	userAgent, err := ReadVarString(buf, 0)
	v.UserAgent = userAgent

	err = readElement(buf, &v.StartHeight)
	if err != nil {
		return err
	}
	err = readElement(buf, &v.Relay)
	if err != nil {
		return err
	}
	return nil

}

func (v *Version) Encode(w io.Writer, protocolVersion uint32) error {
	err := WriteElement(w, ProtocolVersion)
	if err != nil {
		return err
	}

	err = WriteElement(w, DefaultServiceFlag)
	if err != nil {
		return err
	}

	err = WriteElement(w, v.Timestamp)
	if err != nil {
		return err
	}

	err = WriteElement(w, v.Port)
	if err != nil {
		return err
	}

	err = WriteElement(w, v.Nonce)
	if err != nil {
		return err
	}
	err = WriteVarString(w, ProtocolVersion, v.UserAgent)
	if err != nil {
		return err
	}

	err = WriteElement(w, v.StartHeight)
	if err != nil {
		return err
	}

	err = WriteElement(w, v.Relay)
	if err != nil {
		return err
	}

	return nil
}
