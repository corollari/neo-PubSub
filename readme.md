# A simple WebSocket Pub/Sub server for NEO blockchain.

[![Deploy](https://www.herokucdn.com/deploy/button.png)](https://heroku.com/deploy)

#### Under the hood
It spins up a websocket server that connects to a NEO node that is connected to the NEO Network. Whenever the NEO node broadcasts data to its connected clients, this client publishes the data to its subscribed clients by channel. 

#### Dependencies
```
go get github.com/gorilla/websocket
```

#### Run
```bash
go run main.go -network=[main|test]
```

##### Available network
| Network        | Description | Config file
| ------------- |-------------|-------------|
| main      | NEO Main network | config.json |
| test      | NEO Test network | config.testnet.json |


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
| events      | Smart contract events |

**Note**: The node listed on websocketEventsProvider needs to have the NeoPubSub plugin installed along with several other requirements described in [this guide](https://github.com/corollari/neo-node-setupGuide/blob/master/extension-NeoPubSub.md).

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

---

PRs are always welcome.


 
