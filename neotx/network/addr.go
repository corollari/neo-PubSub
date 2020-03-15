package network

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"time"
)

type Endpoint struct {
	IP   net.IP //16 bytes
	Port uint16 //2 bytes
}
type NetWorkAddressWithTime struct {
	Services  ServiceFlag //8 bytes
	Timestamp time.Time   //4 bytes
	Endpoint  Endpoint    //18 bytes. 16 + 2
}

func readNetworkAddress(r io.Reader, na *NetWorkAddressWithTime) error {
	err := readElement(r, (*uint32Time)(&na.Timestamp))
	if err != nil {
		return err
	}
	var ip [16]byte
	err = ReadElements(r, &na.Services, &ip)
	if err != nil {
		return err
	}

	//NOTE: All integer types of NEO are Little Endian except for IP address and port number, these 2 are Big Endian.
	port, err := binarySerializer.Uint16(r, bigEndian)
	if err != nil {
		return err
	}
	//assign to pointer
	*na = NetWorkAddressWithTime{
		Timestamp: na.Timestamp,
		Services:  na.Services,
		Endpoint: Endpoint{
			IP:   net.IP(ip[:]),
			Port: port,
		},
	}
	return nil
}

type Addr struct {
	Addresses []*NetWorkAddressWithTime
}

func (a *Addr) Decode(r io.Reader, protocolVersion uint32) error {
	buf, ok := r.(*bytes.Buffer)

	if !ok {
		return fmt.Errorf("Addr decode reader is not a *byte.Buffer")
	}

	count, err := ReadVarInt(buf, protocolVersion)
	if err != nil {
		return err
	}
	addrList := make([]NetWorkAddressWithTime, count)
	a.Addresses = make([]*NetWorkAddressWithTime, 0, count)
	for i := uint64(0); i < count; i++ {
		na := &addrList[i]
		err := readNetworkAddress(buf, na)
		if err != nil {
			return err
		}
		a.Addresses = append(a.Addresses, na)
	}
	return nil
}

func (a *Addr) Encode(w io.Writer, protocolVersion uint32) error {
	return nil
}
