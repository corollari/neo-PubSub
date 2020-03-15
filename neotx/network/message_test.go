package network

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"testing"
)

func TestNewMessage(t *testing.T) {

	b := NewMessage(NEOMagic, CommandGetAddr, nil)
	fmt.Printf("\nMessage size = %v\n %v", len(b), b)
}

func TestReadMessage(t *testing.T) {

	b := NewMessage(NEOMagic, CommandGetAddr, nil)
	reader := bytes.NewReader(b)

	n, msg, err := ReadMessage(reader, nil)
	fmt.Printf("n = %v\n msg =%v\n err=%v\n", n, msg, err)
}

func TestNewMessageWithVersionpayload(t *testing.T) {
	nonce, _ := RandomUint32()
	payload := NewVersionPayload(10333, nonce)

	b := NewMessage(NEOMagic, CommandVersion, payload)

	reader := bytes.NewReader(b)
	out := &Version{}
	n, msg, err := ReadMessage(reader, out)
	fmt.Printf("n = %v\n msg =%v\n out= %v\n err=%v\n", n, msg, out, err)
}

func TestReadVersionFromNetwork(t *testing.T) {
	// connect to this socket
	conn, err := net.Dial("tcp", "127.0.0.1:20333")

	if err != nil {
		fmt.Println(err)
		return
	}

	//declare version.
	nonce, _ := RandomUint32()
	payload := NewVersionPayload(10333, nonce)
	versionCommand := NewMessage(NEOMagic, CommandVersion, payload)
	conn.Write(versionCommand)

	//read version from peer
	reader := bufio.NewReader(conn)
	out := &Version{}
	n, msg, err := ReadMessage(reader, out)
	fmt.Printf("n = %v\n msg =%v\n out= %v\n err=%v\n", n, msg, out, err)

	//acknowledge the version
	verack := NewMessage(NEOMagic, CommandVerack, nil)
	conn.Write(verack)

	reader2 := bufio.NewReader(conn)
	n2, msg2, err2 := ReadMessage(reader2, nil)
	fmt.Printf("\nn = %v\n msg =%v\n err=%v\n", n2, msg2, err2)

	//getaddress
	getaddr := NewMessage(NEOMagic, CommandGetAddr, nil)
	conn.Write(getaddr)

	reader3 := bufio.NewReader(conn)
	addr := &Addr{}
	n3, msg3, err3 := ReadMessage(reader3, addr)

	fmt.Printf("\nn = %v\n msg =%v\n addr=%v\n err=%v\n", n3, msg3.Command, addr, err3)
	for _, v := range addr.Addresses {
		fmt.Printf("%v\n", v.Endpoint)
	}
}
