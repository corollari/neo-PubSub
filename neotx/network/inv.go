package network

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type InventoryType uint8

const (
	InventotyTypeTX        InventoryType = 0x01 // Transaction = 1
	InventotyTypeBlock     InventoryType = 0x02 // Block = 2
	InventotyTypeConsensus InventoryType = 0xe0 // Consensus data = 224
)

func (i InventoryType) String() string {
	if s, ok := InventoryTypeString[i]; ok {
		return s
	}
	return fmt.Sprintf("Unknown InventoryType: %d", uint32(i))
}

var InventoryTypeString = map[InventoryType]string{
	InventotyTypeTX:        "tx",
	InventotyTypeBlock:     "block",
	InventotyTypeConsensus: "consensus",
}

// NEO Hash Size
const HashSize = 32

type Hash [HashSize]byte

func (h *Hash) ToBytes() []byte {
	b := make([]byte, HashSize)
	for i, j := 0, HashSize-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = byte(h[j]), byte(h[i])
	}
	return b
}

type Inv struct {
	Type   InventoryType //4 bytes
	Hashes []Hash
}

func (v *Inv) Decode(r io.Reader, protocolVersion uint32) error {
	buf, ok := r.(*bytes.Buffer)

	if !ok {
		return fmt.Errorf("Inv decode reader is not a *byte.Buffer")
	}
	//hash is UInt256
	var length uint8
	ReadElements(buf, &v.Type, &length)
	v.Hashes = make([]Hash, length)
	for i := 0; i < int(length); i++ {
		if err := binary.Read(r, binary.LittleEndian, &v.Hashes[i]); err != nil {
			return err
		}
	}
	return nil
}
