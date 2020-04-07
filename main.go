package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/corollari/neo-ws-pub-sub/neotx"
	"github.com/corollari/neo-ws-pub-sub/neotx/network"

	"github.com/corollari/neo-ws-pub-sub/neoutils"
	"github.com/corollari/neo-ws-pub-sub/neorpc"
)

const (
	bufferSize        = 4096
	clientPingPeriod  = 5 * time.Minute

	// Time allowed to write a message to the peer.
	writeWait = 1 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 6 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	serverPingPeriod = (pongWait * 9) / 10
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  16,
	WriteBufferSize: bufferSize,
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

type WebSocketMessage interface{}

type EventMessage struct {
	TxId      string `json:"txid"`
	Contract  string `json:"contract"`
	Event     interface{} `json:"event"`
}

type NodeAddresses struct {
	P2P      string `json:"p2p"`
	RPC      string `json:"rpc"`
}

type Configuration struct {
	Nodes      []NodeAddresses `json:"nodes"`
	WebsocketEventsProvider   string `json:"websocketEventsProvider"`
	Magic         int      `json:"magic"` //network ID.
}

func loadConfigurationFile(file string) (Configuration, error) {
	configuration := Configuration{}
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		return configuration, err
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&configuration)
	return configuration, nil
}

var currentConfig Configuration

//Base on the article 10M Concurrent websocket on https://goroutines.com/10m
func main() {
	mode := flag.String("network", "main", "Network to connect to. main | test")
	portInt := flag.Int("port", 8080, "Port to bind to")
	flag.Parse()

	var file string
	if *mode == "main" {
		file = "config.json"
	} else if *mode == "test" {
		file = "config.testnet.json"
	}

	fmt.Printf("Loading config file:%v\n", file)

	config, err := loadConfigurationFile(file)

	if err != nil {
		fmt.Printf("Error loading config file: %v", err)
		return
	}
	//assign the current configuration to global
	currentConfig = config

	go func() {
		start := time.Now()
		for {
			fmt.Printf("server elapsed=%0.0fs connected=%d failed=%d\n", time.Now().Sub(start).Seconds(), atomic.LoadInt64(&connected), atomic.LoadInt64(&failed))
			time.Sleep(1 * time.Second)
		}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		channel := r.URL.Path[1:]
        if channel[len(channel)-1] == '/' {
            channel = channel[:len(channel)-1] // Remove trailing slash
        }
		contract := r.URL.Query().Get("contract")

		// Close connection if endpoint is not one of the accepted ones
		if channel != "event" && channel != "ping" && channel != "mempool/tx" && channel != "block" {
			http.Error(w, "This endpoint is not available", 404)
			return
		}

		// Upgrade connection to websockets protocol
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		// launch a new goroutine so that this function can return and the http server can free up
		// buffers associated with this connection
		if contract != "" && channel == "event" {
			go handleConnection(ws, contract)
		} else if channel == "ping" {
			go handlePingConnection(ws)
		} else {
			go handleConnection(ws, channel)
		}
	})

	go startConnectToSeed(config, 0)
	go relayEvents(config.WebsocketEventsProvider)

	port := fmt.Sprintf(":%d", *portInt)
	fmt.Printf("Websocket running at port %v\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}

// This endpoint is purposefully undocumented because it was only created for compatibility with neo-mon's latency checks
// TODO: Add deadlines for pings in order to prevent connections being left open?
func handlePingConnection(ws *websocket.Conn) {
	defer ws.Close()
	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			break
		}
		if string(message) == "ping" {
			err = ws.WriteMessage(websocket.TextMessage, []byte("pong"))
			if err != nil {
				break
			}
		}
	}
}

//Handle websocket connection
//Available channels are block, event and mempool/tx.
func handleConnection(ws *websocket.Conn, channel string) {
	sub := subscribe(channel)
	atomic.AddInt64(&connected, 1)
	t := time.NewTicker(clientPingPeriod)

	var message WebSocketMessage
	var ping bool

	for {
		select {
		case <-t.C:
			ping = true
		case message = <-sub:
			ping = false
		}

		ws.SetWriteDeadline(time.Now().Add(30 * time.Second))
		var err error
		if ping == true {
			err = ws.WriteMessage(websocket.PingMessage, nil)
		} else {
			err = ws.WriteJSON(message)
		}
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

// Adapted from https://github.com/gorilla/websocket/blob/master/examples/echo/client.go
// Ping/pong system based on https://github.com/gorilla/websocket/blob/master/examples/chat/client.go
func relayEvents(WebsocketEventsProvider string) {
	log.Printf("connecting to %s", WebsocketEventsProvider)

	c, _, err := websocket.DefaultDialer.Dial(WebsocketEventsProvider, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	// Restart connection when lost
	defer func() {
		log.Printf("connection to %s lost, reconnecting...", WebsocketEventsProvider)
		// TODO: Set up exponential back-off system
		time.Sleep(10 * time.Second)
		go relayEvents(WebsocketEventsProvider)
	}()

	done := make(chan struct{})

	go func() {
		defer close(done)

		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}

			go broadcastMessage(message)
		}
	}()

	// Set up ping system
	c.SetReadDeadline(time.Now().Add(pongWait))
	c.SetPongHandler(func(string) error { c.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	ticker := time.NewTicker(serverPingPeriod) // Ping timer
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			c.SetWriteDeadline(time.Now().Add(writeWait))
			err := c.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				log.Println("Error on ping write:", err)
				return
			}
		}
	}
}

func broadcastMessage(message []byte) {
	var decodedMessage map[string]interface{}
	if err := json.Unmarshal(message, &decodedMessage); err != nil {
        panic(err)
    }

    msgType := decodedMessage["type"].(string)
    msgPayload := decodedMessage["data"].(map[string]interface{})

    if msgType == "events" {
        contract := msgPayload["contract"].(string)
        log.Printf("received event on %s", contract)

        m := EventMessage {
            msgPayload["txid"].(string),
            contract,
            msgPayload["call"].(map[string]interface{})["value"],
        }

        sendMessage("event", m)
        sendMessage(contract, m)
    } else if msgType == "blocks" {
        sendMessage("block", msgPayload)
    }
}

func getBestNode(list []string) *neoutils.SeedNodeResponse {
	commaSeparated := strings.Join(list, ",")
	return neoutils.SelectBestSeedNode(commaSeparated)
}

var connectedToNEONode = false

//this is NEO part
func startConnectToSeed(config Configuration, iteration int) {
	node := config.Nodes[iteration]
	host, port, err := net.SplitHostPort(node.P2P)
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(-1)
	}
	portInt, err := strconv.Atoi(port)
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(-1)
	}
	var neoNodeConfig = neotx.Config{
		Network:   network.NEONetworkMagic(config.Magic),
		Port:      uint16(portInt),
		IPAddress: host,
	}
	client := neotx.NewClient(neoNodeConfig)
	handler := &NEOConnectionHandler{}
	handler.config = config
	handler.nodeNumber = iteration

	client.SetDelegate(handler)

	fmt.Printf("connecting to %v:%v...\n", neoNodeConfig.IPAddress, neoNodeConfig.Port)
	err = client.Start()
	if err != nil {
		if iteration < len(config.Nodes){
			go startConnectToSeed(config, iteration + 1)
		} else {
			fmt.Printf("%v", err)
			os.Exit(-1)
		}
	}
}

type NEOConnectionHandler struct {
	config Configuration
	nodeNumber int
}

//implement the message protocol
func (h *NEOConnectionHandler) OnReceive(tx neotx.TX) {

	if tx.Type == network.InventotyTypeTX {
		//Call getrawtransaction to get the transaction detail by txid

		rpcNode := h.config.Nodes[h.nodeNumber].RPC //Has to communicate with the same node that it got the transaction from, otherwise another node might not know about the tx

		client := neorpc.NewClient(rpcNode)
		raw := client.GetRawTransaction(tx.ID)

		if raw.ErrorResponse != nil {
			return
		}

		m := raw.Result

		fmt.Printf(" %v: %+v", tx.ID, raw.Result.Type)
		sendMessage("mempool/tx", m)
		return
	}
	// The remaining type of inv message are consensus and block, we ignore them
}

func (h *NEOConnectionHandler) OnConnected(c network.Version) {
	fmt.Printf("connected %+v\n", c)
	connectedToNEONode = true
}

func (h *NEOConnectionHandler) OnError(e error) {
	fmt.Printf("Disconnected from host. Trying to connect to a different host...")
	go startConnectToSeed(currentConfig, (h.nodeNumber + 1) % len(h.config.Nodes))
}
