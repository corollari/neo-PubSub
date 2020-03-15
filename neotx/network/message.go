package network

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"io"
)

// http://docs.neo.org/en-us/node/network-protocol.html

// A network protocol ID
// Production mode = 7630401
type NEONetworkMagic uint32

const (
	// NEO message header is Magic(4) + Command(12) + Payload Legth(4) + Checksum(4)
	MessageHeaderSize = 24
	MagicSize         = 4
	CommandSize       = 12 // Fixed size for command. Shorter command must be zero padded
	PayloadLegthSize  = 4
	CheckSumSize      = 4
	MaxMessagePayload = (1024 * 1024 * 10) // 10MB
)

type Command string

const (
	CommandAddr        Command = "addr"
	CommandBlock       Command = "block"
	CommandConsensus   Command = "consensus"
	CommandFilterAdd   Command = "filteradd"
	CommandFilterClear Command = "filterclear"
	CommandFilterLoad  Command = "filterload"
	CommandGetAddr     Command = "getaddr"
	CommandGetBlocks   Command = "getblocks"
	CommandGetData     Command = "getdata"
	CommandGetHeaders  Command = "getheaders"
	CommandHeaders     Command = "headers"
	CommandMempool     Command = "mempool"
	CommandTx          Command = "tx"
	CommandVerack      Command = "verack"
	CommandVersion     Command = "version"
	CommandInv         Command = "inv"

	//no operation yet
	CommandAlert       Command = "alert"
	CommandMerkleblock Command = "merkleblock"
	CommandNotfound    Command = "notfound"
	CommandPing        Command = "ping"
	CommandPong        Command = "pong"
	CommandReject      Command = "reject"
)

func getChecksum(payload []byte) uint32 {
	first := sha256.Sum256(payload)
	second := sha256.Sum256(first[:])
	//checksum is first 4 bytes
	return binary.LittleEndian.Uint32(second[0:4])
}

func NewMessage(magic NEONetworkMagic, command Command, p PayloadInterface) []byte {
	payload := []byte{}
	if p != nil {
		var bw bytes.Buffer
		p.Encode(&bw, ProtocolVersion)
		payload = bw.Bytes()
	}
	buffer := make([]byte, MessageHeaderSize+len(payload))
	checksum := getChecksum(payload)
	binary.LittleEndian.PutUint32(buffer, uint32(magic))             //fist 4 bytes
	copy(buffer[4:], []byte(command))                                //from 4th byte
	binary.LittleEndian.PutUint32(buffer[16:], uint32(len(payload))) //from 16th byte
	binary.LittleEndian.PutUint32(buffer[20:], checksum)             //from 20th byte

	if len(payload) > 0 {
		copy(buffer[24:], payload) //from 24th byte
	}
	return buffer
}

type MessageHeader struct {
	Magic    NEONetworkMagic // 4 bytes
	Command  string          // 12 bytes
	Length   uint32          // 4 bytes
	Checksum [4]byte         // 4 bytes. Int32
	Payload  []byte          // Length bytes
}

func ReadMessage(r io.Reader, payloadOutput PayloadInterface) (int, *MessageHeader, error) {

	var headerBytes [MessageHeaderSize]byte

	//ReadFull reads exactly len(buf) bytes from r into buf. It returns the number of bytes copied and an error if fewer
	//bytes were read. The error is EOF only if no bytes were read. If an EOF happens after reading some but not all
	//the bytes, ReadFull returns ErrUnexpectedEOF. On return, n == len(buf) if and only if err == nil.
	n, err := io.ReadFull(r, headerBytes[:])
	if err != nil {
		return n, nil, err
	}

	reader := bytes.NewReader(headerBytes[:])

	message := MessageHeader{}
	var command [CommandSize]byte
	ReadElements(reader, &message.Magic, &command, &message.Length, &message.Checksum)
	//Remove trailing zero
	message.Command = string(bytes.TrimRight(command[:], string(0)))

	totalBytes := 0
	totalBytes += n

	//TODO validation, check size
	if payloadOutput != nil {
		payloadByte := make([]byte, message.Length)
		n, err = io.ReadFull(r, payloadByte)
		if err != nil {
			return n, nil, err
		}
		totalBytes += n
		pr := bytes.NewBuffer(payloadByte)
		payloadOutput.Decode(pr, 0)
		return totalBytes, &message, nil
	}

	return n, &message, nil
}
