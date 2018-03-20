package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/o3labs/neo-transaction-watcher/neotx"
	"github.com/o3labs/neo-transaction-watcher/neotx/network"

	"github.com/o3labs/neo-utils/neoutils/neorpc"
	"github.com/o3labs/neo-utils/neoutils/smartcontract"
)

//Settings here
const (
	websocketPort      = 8080
	neoJSONRPCEndpoint = "http://seed1.o3node.org:10332"
)

var neoNodeConfig = neotx.Config{
	Network:   neotx.NEOMainNet,
	Port:      10333,
	IPAddress: "seed1.o3node.org",
}

const (
	maxMessageSize = 256
	pingPeriod     = 5 * time.Minute
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  16,
	WriteBufferSize: maxMessageSize,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	subscriptions      = map[string][]chan WebSocketMessage{}
	subscriptionsMutex sync.Mutex
)

var (
	connected int64
	failed    int64
)

type WebSocketMessage struct {
	Type string      `json:"type"`
	TXID string      `json:"txID"`
	Data interface{} `json:"data,omitempty"`
}

//Base on the article 10M Concurrent webcoket on https://goroutines.com/10m
func main() {
	go func() {
		start := time.Now()
		for {
			fmt.Printf("server elapsed=%0.0fs connected=%d failed=%d\n", time.Now().Sub(start).Seconds(), atomic.LoadInt64(&connected), atomic.LoadInt64(&failed))
			time.Sleep(1 * time.Second)
		}
	}()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		channel := r.URL.Query().Get("type")
		// launch a new goroutine so that this function can return and the http server can free up
		// buffers associated with this connection
		go handleConnection(ws, channel)
	})

	go startConnectToSeed()

	port := fmt.Sprintf(":%d", websocketPort)
	fmt.Printf("Websocket running at port %v\n", websocketPort)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}

}

func handleConnection(ws *websocket.Conn, channel string) {
	sub := subscribe(channel)
	atomic.AddInt64(&connected, 1)
	t := time.NewTicker(pingPeriod)

	var message WebSocketMessage

	for {
		select {
		case <-t.C:
			message = WebSocketMessage{}
		case message = <-sub:
		}

		ws.SetWriteDeadline(time.Now().Add(30 * time.Second))
		err := ws.WriteJSON(message)
		if err != nil {
			break
		}
	}
	atomic.AddInt64(&connected, -1)
	atomic.AddInt64(&failed, 1)

	t.Stop()
	ws.Close()
	unsubscribe(channel, sub)
}

func subscribe(channel string) chan WebSocketMessage {
	sub := make(chan WebSocketMessage)
	subscriptionsMutex.Lock()
	subscriptions[channel] = append(subscriptions[channel], sub)
	subscriptionsMutex.Unlock()
	return sub
}

func unsubscribe(channel string, sub chan WebSocketMessage) {
	subscriptionsMutex.Lock()
	newSubs := []chan WebSocketMessage{}
	subs := subscriptions[channel]
	for _, s := range subs {
		if s != sub {
			newSubs = append(newSubs, s)
		}
	}
	subscriptions[channel] = newSubs
	subscriptionsMutex.Unlock()
}

func sendMessage(channel string, message WebSocketMessage) {
	subscriptionsMutex.Lock()
	subs := subscriptions[channel]
	subscriptionsMutex.Unlock()
	for _, s := range subs {
		select {
		case s <- message:
		default:
			// drop the message if nobody is ready to receive it
		}
	}
}

//this is NEO part
func startConnectToSeed() {
	err := neotx.Start(neoNodeConfig, onReceivedTX)
	if err != nil {
		log.Printf("%v", err)
		os.Exit(-1)
	}
}

func onReceivedTX(tx neotx.TX) {

	if tx.Type == network.InventotyTypeTX {
		//Call getrawtransaction to get the transaction detail by txid

		client := neorpc.NewClient(neoJSONRPCEndpoint)
		raw := client.GetRawTransaction(tx.ID)

		if raw.ErrorResponse != nil {
			return
		}
		m := WebSocketMessage{
			Type: tx.Type.String(),
			TXID: tx.ID,
			Data: raw,
		}
		log.Printf(" %v: %+v", tx.ID, raw.Result.Type)
		if raw.Result.Type == "InvocationTransaction" {
			parser := smartcontract.NewParserWithScript(raw.Result.Script)
			result, err := parser.GetListOfScriptHashes()
			if err != nil {
				return
			}
			for _, v := range result {

				log.Printf("result = %v\n", v)
			}
		}
		sendMessage(tx.Type.String(), m)
		return
	}
	//another type of INV. consensus and block
	m := WebSocketMessage{
		Type: tx.Type.String(),
		TXID: tx.ID,
	}
	log.Printf("%+v", m)
	sendMessage(tx.Type.String(), m)

}
