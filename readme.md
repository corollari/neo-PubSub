# A simple WebSocket Pub/Sub server for NEO blockchain.

[![Deploy](https://www.herokucdn.com/deploy/button.png)](https://heroku.com/deploy)

#### Under the hood
It spins up a websocket server and uses [neo-transaction-watcher](https://github.com/O3Labs/neo-transaction-watcher) to connect to a NEO node that is connected to the NEO Network. whenever the NEO node broadcasts data to its connected clients. This client publishes the data to its subscribed clients by channel. 

#### Dependencies
```
go get github.com/gorilla/websocket
go get github.com/corollari/neo-transaction-watcher
go get github.com/corollari/neo-utils/neoutils
```

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

### Heroku deployment
```bash
# See https://elements.heroku.com/buildpacks/heroku/heroku-buildpack-multi-procfile
MAINNETAPP=neo-ws-pub-sub-mainnet
TESTNETAPP=neo-ws-pub-sub-testnet
heroku create -a $MAINNETAPP --region eu
heroku create -a $TESTNETAPP --region eu
heroku buildpacks:add -a $MAINNETAPP https://github.com/heroku/heroku-buildpack-multi-procfile
heroku buildpacks:add -a $MAINNETAPP heroku/go
heroku buildpacks:add -a $TESTNETAPP https://github.com/heroku/heroku-buildpack-multi-procfile
heroku buildpacks:add -a $TESTNETAPP heroku/go
heroku config:set -a $MAINNETAPP PROCFILE=ProcfileMainnet
heroku config:set -a $TESTNETAPP PROCFILE=ProcfileTestnet
git push https://git.heroku.com/$MAINNETAPP.git HEAD:master
git push https://git.heroku.com/$TESTNETAPP.git HEAD:master
```

Afterwards the following changes can be made manually from heroku's dashboard:
- Change dyno type from Free to Hobby in order to get continuous uptime & SSL certificates
- Connect the heroku apps to a github repo in order to enable automatic deploys

#### Example
See `example/client.html`

#### TODO
- [ ] subscribe to particular smart contract.  
- [ ] subscribe to particular function in a smart contract.  

---

PRs are always welcome.


 
