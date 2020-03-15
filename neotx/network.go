package neotx

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/corollari/neo-ws-pub-sub/neotx/network"
)

//Network
const (
	NEOMainNet    network.NEONetworkMagic = 7630401
	NEOTestNet    network.NEONetworkMagic = 1953787457
	NEOPrivateNet network.NEONetworkMagic = 56753
)

type Config struct {
	Network   network.NEONetworkMagic
	IPAddress string
	Port      uint16
}

type Client struct {
	Config     Config
	delegate   MessageDelegate
	connection net.Conn
}

func NewClient(config Config) *Client {
	return &Client{Config: config}
}

type TX struct {
	Type network.InventoryType
	ID   string
}

type MessageDelegate interface {
	OnReceive(TX)
	OnConnected(network.Version)
	OnError(error)
}

type Interface interface {
	handleConnection()
	Start() error
	SetDelegate(MessageDelegate)
}

func startPingLoop(c *Client) {
	conn := c.connection
	for {
		time.Sleep(1000 * time.Millisecond)
		nonce, _ := network.RandomUint32()
		payload := network.NewPingPayload(nonce)
		pingCommand := network.NewMessage(c.Config.Network, network.CommandPing, payload)
		conn.Write(pingCommand)
	}
}

var _ Interface = (*Client)(nil)

func (c *Client) handleConnection() {
	conn := c.connection
	log.Printf("remote address = %v", conn.RemoteAddr().String())
	log.Printf("local address = %v", conn.LocalAddr().String())
	nonce, _ := network.RandomUint32()
	payload := network.NewVersionPayload(c.Config.Port, nonce)
	versionCommand := network.NewMessage(c.Config.Network, network.CommandVersion, payload)
	conn.Write(versionCommand)

	for {
		_, msg, err := network.ReadMessage(conn, nil)
		if err != nil {
			log.Printf("mesage from server when error %+v", err)
			if c.delegate != nil {
				c.connection.Close()
				go c.delegate.OnError(err)
			}
			return
		}

		payloadByte := make([]byte, msg.Length)
		_, err = io.ReadFull(conn, payloadByte)
		if err != nil {
			if c.delegate != nil {
				go c.delegate.OnError(err)
			}
			continue
		}

		//receive version from remote node
		if msg.Command == string(network.CommandVersion) {
			out := &network.Version{}
			pr := bytes.NewBuffer(payloadByte)
			out.Decode(pr, 0)
			//reply with verack
			verack := network.NewMessage(c.Config.Network, network.CommandVerack, nil)
			conn.Write(verack)

			if c.delegate != nil {
				go c.delegate.OnConnected(*out)
			}
		} else if msg.Command == string(network.CommandVerack) {
			go startPingLoop(c)
		} else if msg.Command == string(network.CommandPong) {
			log.Println("Pong!")
		} else if msg.Command == string(network.CommandAddr) {
			out := &network.Addr{}
			pr := bytes.NewBuffer(payloadByte)
			out.Decode(pr, 0)
		} else if msg.Command == string(network.CommandInv) {
			out := &network.Inv{}
			log.Printf("msg = %+v\n", msg)
			pr := bytes.NewBuffer(payloadByte)
			out.Decode(pr, 0)
			for _, v := range out.Hashes {
				b := v.ToBytes()
				txID := hex.EncodeToString(b)
				tx := TX{
					Type: out.Type,
					ID:   txID,
				}

				if c.delegate != nil {
					go c.delegate.OnReceive(tx)
				}
			}
		} else {
			log.Printf("Unhandled message: %s\n", msg.Command)
		}
	}
}
func (c *Client) SetDelegate(d MessageDelegate) {
	c.delegate = d
}

func (c *Client) Start() error {

	address := fmt.Sprintf("%v:%v", c.Config.IPAddress, c.Config.Port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}

	c.connection = conn
	c.handleConnection()
	return nil
}
