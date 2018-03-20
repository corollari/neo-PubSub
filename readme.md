# A simple WebSocket Pub/Sub server for NEO blockchain.


#### Under the hood
It spins up a websocket server and uses [neo-transaction-watcher](https://github.com/O3Labs/neo-transaction-watcher) to connect to a NEO node that is connected to the NEO Network. whenever the NEO node broadcasts data to its connected clients. This client publishes the data to its subscribed clients by channel. 

#### Run
```bash
go run main.go -network=[main|test|private]
```

##### Available network
| Network        | Description | Config file
| ------------- |-------------|-------------|
| main      | NEO Main network | config.json |
| test      | NEO Test network | config.testnet.json |
| private      | NEO private network running locally | config.privatenet.json |


#### Websocket URL
```
ws://localhost:8080/?channel=tx
```

##### Available channels
| Channel        | Description |
| ------------- |-------------|
| consensus      | Consensus data |
| block      | Block |
| tx      | Transaction |


#### Example
See `example/client.html`

#### TODO
- [ ] subscribe to particular smart contract.  
- [ ] subscribe to particular function in a smart contract.  

---

PRs are always welcome.


 
