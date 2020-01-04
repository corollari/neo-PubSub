package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
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
	"github.com/corollari/neo-transaction-watcher/neotx"
	"github.com/corollari/neo-transaction-watcher/neotx/network"

	"github.com/corollari/neo-ws-pub-sub/neoutils"
	"github.com/corollari/neo-ws-pub-sub/neorpc"
)

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

type Configuration struct {
	WebsocketPort uint     `json:"websocketPort"`
	SeedList      []string `json:"seedList"`
	RPCSeedList   []string `json:"rpcSeedList"`
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
	mode := flag.String("network", "", "Network to connect to. main | test | private")
	portInt := flag.Int("port", 8080, "Port to bind to")
	flag.Parse()

	if *mode == "" {
		//default mode is private
		defaultEnv := "main"
		mode = &defaultEnv
	}

	file := "config.privatenet.json"
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
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		channel := r.URL.Query().Get("channel")
		// launch a new goroutine so that this function can return and the http server can free up
		// buffers associated with this connection
		go handleConnection(ws, channel)
	})

	go startConnectToSeed(config)

	port := fmt.Sprintf(":%d", *portInt)
	fmt.Printf("Websocket running at port %v\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}

//Handle websocket connection
//Available channels are consensus,block and tx.
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

func getBestNode(list []string) *neoutils.SeedNodeResponse {
	commaSeparated := strings.Join(list, ",")
	return neoutils.SelectBestSeedNode(commaSeparated)
}

var connectedToNEONode = false

//this is NEO part
func startConnectToSeed(config Configuration) {
	first := config.SeedList[0]
	host, port, err := net.SplitHostPort(first)
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

	client.SetDelegate(handler)

	fmt.Printf("connecting to %v:%v...\n", neoNodeConfig.IPAddress, neoNodeConfig.Port)
	err = client.Start()
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(-1)
	}
}

type NEOConnectionHandler struct {
	config Configuration
}

//implement the message protocol
func (h *NEOConnectionHandler) OnReceive(tx neotx.TX) {

	if tx.Type == network.InventotyTypeTX {
		//Call getrawtransaction to get the transaction detail by txid

		best := getBestNode(h.config.RPCSeedList)
		if best == nil {
			return
		}

		client := neorpc.NewClient(best.URL)
		raw := client.GetRawTransaction(tx.ID)

		if raw.ErrorResponse != nil {
			return
		}
		m := WebSocketMessage{
			Type: tx.Type.String(),
			TXID: tx.ID,
			Data: raw,
		}
		fmt.Printf(" %v: %+v", tx.ID, raw.Result.Type)
		// if raw.Result.Type == "InvocationTransaction" {
		// 	parser := smartcontract.NewParserWithScript(raw.Result.Script)
		// 	result, err := parser.GetListOfScriptHashes()
		// 	if err != nil {
		// 		return
		// 	}
		// 	for _, v := range result {

		// 		fmt.Printf("result = %v\n", v)
		// 	}
		// }
		sendMessage(tx.Type.String(), m)
		return
	}
	//another type of INV. consensus and block
	m := WebSocketMessage{
		Type: tx.Type.String(),
		TXID: tx.ID,
	}
	fmt.Printf("%+v", m)
	sendMessage(tx.Type.String(), m)
}

func (h *NEOConnectionHandler) OnConnected(c network.Version) {
	fmt.Printf("connected %+v\n", c)
	connectedToNEONode = true
}

func (h *NEOConnectionHandler) OnError(e error) {
	if e == io.EOF && connectedToNEONode == true {
		connectedToNEONode = false
		fmt.Printf("Disconnected from host. will try to connect in 15 seconds...")
		for {
			time.Sleep(15 * time.Second)
			//we need to implement backoff and retry to reconnect here
			//if the error is EOF then we try to reconnect
			go startConnectToSeed(currentConfig)
		}
	}
}
