package network

import (
	"bytes"
	"log"
	"testing"
)

func TestVersionMessage(t *testing.T) {
	payload := NewVersionPayload(10333, uint32(25505))
	log.Printf("%v", payload)
}

func TestEncode(t *testing.T) {
	nonce, _ := RandomUint32()
	p := NewVersionPayload(10333, nonce)
	var payload bytes.Buffer
	p.Encode(&payload, uint32(0))
	log.Printf("%v %v", payload.Bytes(), len(payload.Bytes()))

	var receivedVersion Version
	receivedVersion.Decode(&payload, 0)
	log.Printf("%v", receivedVersion)
}
