package main

import (
	"log"
	"os"

	"github.com/corollari/neo-ws-pub-sub/neo-transaction-watcher/neotx"
	"github.com/corollari/neo-ws-pub-sub/neo-transaction-watcher/neotx/network"
)

type Handler struct {
}

//implement the message protocol
func (h *Handler) OnReceive(tx neotx.TX) {
	log.Printf("%+v", tx)
}

func (h *Handler) OnConnected(c network.Version) {
	log.Printf("connected %+v", c)
}

func (h *Handler) OnError(e error) {
	log.Printf("error %+v", e)
}

func main() {
	config := neotx.Config{
		Network:   neotx.NEOMainNet,
		Port:      10333,
		IPAddress: "52.77.48.175",
	}
	client := neotx.NewClient(config)
	handler := &Handler{}
	client.SetDelegate(handler)

	err := client.Start()
	if err != nil {
		log.Printf("%v", err)
		os.Exit(-1)
	}

	for {

	}
}
