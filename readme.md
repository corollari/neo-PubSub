![banner](https://raw.githubusercontent.com/corollari/neo-PubSub/master/website/banner.png)

<h4 align="center">A simple WebSocket Pub/Sub server for the NEO blockchain</h4>

<p align="center">
  <a href="#use">Use</a> •
  <a href="#build">Build</a> •
  <a href="#deploy">Deploy</a> •
  <a href="#example-events">Example events</a> •
  <a href="#license">License</a>
</p>

## Use
We built a publish-subscribe service for NEO that allows anyone to listen for events happening on the blockchain and act on them. The service is based in websockets and can be used by just connecting to the following endpoints:
```
wss://pubsub.main.neologin.io?channel=events # Events triggered in smart contract executions (final)
wss://pubsub.main.neologin.io?channel=tx # A transaction has entered in the mempool (but may not be inside a block yet)
wss://pubsub.main.neologin.io?channel=block # A block is propagated (not finalized tho)
```

You can test these websocket endpoints directly [here](https://corollari.github.io/neo-PubSub/) or check [some response examples](#example-events).

**Note**: If you wish to get events from testnet instead of mainnet just replace the `main` prefix with `test`, eg: `pubsub.test.neologin.io`.

Using javascript, you can connect to these endpoints with the following code:
```js
const ws = new WebSocket('wss://pubsub.main.neologin.io/?channel=events')
ws.onmessage = (event) => {
  console.log(event.data);
}
```

You can also check the state of the service at [CoZ Monitor](http://monitor.cityofzion.io/).

Going forward we will continue to host and maintain these servers for the community to use.

### How does it work
The server captures messages sent through the p2p network that connects all the NEO nodes and relays them to the clients. The `events` channel is different because instead of deriving the data from the p2p network it gets the events from a node that has the NeoPubSub plugin installed, capturing all the event calls triggered in successful contract executions by transactions that have been included in a final block.

### Scalability
We've designed the system to handle high load and be able to scale horizontally. This is possible thanks to an architecture based around a singleton instance of a NEO node that gets all the data which is then relayed to a scalable amount of dynos hosted on heroku which maintain the websockets connections with all the clients and push the data to them, thus we can dynamically scale the amount of dynos on heroku to meet demand.

## Build
```bash
go get # Install dependencies
go run main.go -network=[main|test] # Run server
```

When it's running, you can use it by connecting to the following endpoint:
```
ws://localhost:8080/?channel=tx
```

##### Available channels
| Channel        | Description |
| ------------- |-------------|
| block      | Block |
| tx      | Transaction |
| events      | Smart contract events |

### Available networks
| Network        | Description | Config file
| ------------- |-------------|-------------|
| main      | NEO Main network | config.json |
| test      | NEO Test network | config.testnet.json |

**Note**: The node listed on the websocketEventsProvider field needs to have the NeoPubSub plugin installed along with several other requirements described in [this guide](https://github.com/corollari/neo-node-setupGuide/blob/master/extension-NeoPubSub.md).

## Deploy
Deployment to heroku as two different apps (one for mainnet and the other one for testnet) can be done using the following steps:

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

Afterwards, the following changes can be made manually from heroku's dashboard:
- Change dyno type from Free to Hobby in order to get continuous uptime & SSL certificates
- Connect the heroku apps to a github repo in order to enable automatic deploys

## Example events
### tx
```json
{
   "type":"tx",
   "txID":"cb11e90091069b5086b723326e15181e717343e7be49f90709c934f328beea0d",
   "data":{
      "jsonrpc":"2.0",
      "id":1,
      "result":{
         "txid":"0xcb11e90091069b5086b723326e15181e717343e7be49f90709c934f328beea0d",
         "size":215,
         "type":"ContractTransaction",
         "version":0,
         "attributes":[
            {
               "usage":"Remark",
               "data":"4f3358464f52434c41494d"
            }
         ],
         "vin":[
            {
               "txid":"0x3bd33d4183f9e207be9308b42a489e595a6fe6099f23ac7bfff2ee65a99087a8",
               "vout":0
            }
         ],
         "vout":[
            {
               "n":0,
               "asset":"0xc56f33fc6ecfcd0c225c4ab356fee59390af8560be0e930faebe74a6daff7c9b",
               "value":"254",
               "address":"AXmT1qBWTvzu434eBrVGx9VwvPhLCZapA6"
            }
         ],
         "claims":null,
         "sys_fee":"0",
         "net_fee":"0",
         "scripts":[
            {
               "invocation":"4050e76610d39cdce6b31fd07ad5bb5009bf09966d55fa510d828cb2efcd0b415e3c5700db1ea414695b3f1c1ebc5c3cffbce9792a29463e178d99d88f76468343",
               "verification":"2102f9bec8e6da87dd4af85f76030847e4f0e3ea467ef24f9fd2077c495f6f91b458ac"
            }
         ],
         "script":"",
         "gas":"",
         "blockhash":"",
         "confirmations":0,
         "blocktime":0
      }
   }
}
```

### block
```json
{
   "type":"block",
   "txID":"110603a66d2353271506d255d3204caf330850b58b39950c87282fd7aafb0554"
}
```

### events
```json
{
   "type":"events",
   "txID":"0x92c6e0510d23cb9c12d52f7494e88be07a0bf139c126d8804ed90348714f165c",
   "data":{
      "contract":"0xfde69a7dd2a1c948977fb3ce512158987c0e2197",
      "call":{
         "type":"Array",
         "value":[
            {
               "type":"ByteArray",
               "value":"6f7261636c654f70657261746f72"
            },
            {
               "type":"ByteArray",
               "value":"55d6d86bdec15db437aca45b4e8705333f1fdb07"
            },
            {
               "type":"ByteArray",
               "value":"736e656f5f7072696365"
            },
            {
               "type":"ByteArray",
               "value":""
            },
            {
               "type":"ByteArray",
               "value":"b0273755"
            },
            {
               "type":"Integer",
               "value":"2"
            }
         ]
      }
   }
}
```

## License
Apache License 2.0

The initial version of this package was built by the O3 team.

-----

Built with ❤️ by the team behind [NeoLogin: a simple and easy to integrate wallet provider for Neo dApps](https://neologin.io).
