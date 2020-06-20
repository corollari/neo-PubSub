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
wss://pubsub.main.neologin.io/event # Events triggered in smart contract executions (final)
wss://pubsub.main.neologin.io/event?contract=0xfb84b0950e8fd366af566b2911d6183e4b0367f7 # Events from a specific contract
wss://pubsub.main.neologin.io/mempool/tx # A transaction has entered in the mempool (but may not be inside a block yet)
wss://pubsub.main.neologin.io/block # A block is finalized
```

You can test these websocket endpoints directly [here](https://corollari.github.io/neo-PubSub/) or check [some response examples](#example-events).

**Note**: If you wish to get events from testnet instead of mainnet just replace the `main` prefix with `test`, eg: `pubsub.test.neologin.io`.

Using javascript, you can connect to these endpoints with the following code:
```js
const ws = new WebSocket('wss://pubsub.main.neologin.io/event')
ws.onmessage = (event) => {
  console.log(event.data);
}
```

Or, if you use `neon-js`, it's also possible to subscribe to events in the following way:
```js
import { api } from "@cityofzion/neon-js";
const mainNetNotifications = new api.notifications.instance("MainNet");
const subscription = mainNetNotifications.subscribe("0x314b5aac1cdd01d10661b00886197f2194c3c89b", (event) => {
  console.log(event); // Print the events being received in real time
});
```

You can also check the state of the service at [CoZ Monitor](http://monitor.cityofzion.io/).

Going forward we will continue to host and maintain these servers for the community to use.

### How does it work
The server captures messages sent through the p2p network that connects all the NEO nodes and relays them to the clients. The `event` channel is different because instead of deriving the data from the p2p network it gets the events from a node that has the NeoPubSub plugin installed, capturing all the event calls triggered in successful contract executions by transactions that have been included in a final block.

### Scalability
We've designed the system to handle high load and be able to scale horizontally. This is possible thanks to an architecture based around a singleton instance of a NEO node that gets all the data which is then relayed to a scalable amount of dynos hosted on heroku which maintain the websockets connections with all the clients and push the data to them, thus we can dynamically scale the amount of dynos on heroku to meet demand.

## Build
```bash
go get # Install dependencies
go run main.go -network=[main|test] # Run server
```

When it's running, you can use it by connecting to the following endpoint:
```
ws://localhost:8080/event
```

##### Available channels
| Channel        | Description |
| ------------- |-------------|
| event      | Smart contract event |
| block      | Block |
| mempool/tx      | Mempool Transaction |

The `event` channel can be filtered by contract with the query parameter `contract`. For example, `wss://pubsub.main.neologin.io/event?contract=0xfb84b0950e8fd366af566b2911d6183e4b0367f7` will only receive events triggered inside the `0xfb84b0950e8fd366af566b2911d6183e4b0367f7` contract.

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
### event
```json
{
   "event":[
      {
         "type":"ByteArray",
         "value":"7472616e73666572"
      },
      {
         "type":"ByteArray",
         "value":"30074a2d88bab26f74142c188231e92ad401dbf6"
      },
      {
         "type":"ByteArray",
         "value":"8ba6205856117b0f3909cd88209aa919ec9c14b8"
      },
      {
         "type":"ByteArray",
         "value":"00c39dd000"
      }
   ],
   "contract":"0x314b5aac1cdd01d10661b00886197f2194c3c89b",
   "txid":"0xd6f5185a19abad3f3bbea88ac4ec63b449ac38908bd7761dce75e445502bc76f"
}
```

### mempool/tx
```json
{
   "txid":"0x70a002b957ff29824b84222c7f6550694421d3445ebc1ce8778864cae5760817",
   "size":228,
   "type":"InvocationTransaction",
   "version":1,
   "attributes":[
      {
         "usage":"Script",
         "data":"ac5402b4fb3e02bba42506c051e07e64e39f9ce0"
      }
   ],
   "vin":[

   ],
   "vout":[

   ],
   "claims":null,
   "sys_fee":"0",
   "net_fee":"0",
   "scripts":[
      {
         "invocation":"40366afbfdc2437e6f8d7fe975d7fdc10bed151d2e8d38cd524474d6a3754e181220c8877505be96ed91ad54317a1978605b4d27975076ca8b76a7795616c4ff80",
         "verification":"21038666b29b1f87d6b797d40b96dd5f18e1851d6a10e4db26252b8069ca952c4326ac"
      }
   ],
   "script":"05000d290704144649f940788c16c01e70d39e3613252ef7a58dbc14ac5402b4fb3e02bba42506c051e07e64e39f9ce053c1087472616e7366657267f767034b3e18d611296b56af66d38f0e95b084fbf166514edb030272a89f",
   "gas":"0",
   "blockhash":"",
   "confirmations":0,
   "blocktime":0
}
```

### block
```json
{
   "confirmations":1,
   "hash":"0x715c921fa65352b657afd8db82a1e65d7ea0cf6686fc30f3bf80a607cc6fff4d",
   "index":5249790,
   "merkleroot":"0x63b87f72f4d77aa82de6dc19523d07abb4300136d79e9a9c2b251d0eec01b8ce",
   "nextconsensus":"ANuupE2wgsHYi8VTqSUSoMsyxbJ8P3szu7",
   "nonce":"8fee67b29ce528aa",
   "previousblockhash":"0xf580c47649b3c22d1699dfcad8b396458ee4572b79b6faba5ca3c14493e94f42",
   "script":{
      "invocation":"40e3d5b93963011fce63ce0529dd105a87395babea4f5d5840110113a677584c99f51ce57196204bcbd73393a72998278bfdc30540b10d7a9e5e342b4265aea9d340bc2b3992246e0049645685d85b29b757ec3b5b0e249c038cd18ad682f8e18d438c9ff2b89cf63bd94cf26a6d16f23c9b377ad03c8fd07d076a73b9e801ab9c07406e254f902d5f648b68a7c0652bcf9b57b5dadb7c6d2d159514bde83b90e30433246efdc81e612c6ad1d8fab2f73c3382dccdc5d09536b1342f70a2d6493df5da40c0d3cfd4d3453e8e4c78cf53978e6adaf0f6bdf071bec8b912a9c8e5ff51f5fcf046096d8b2a88558cb5fe32371c709136a9b99624e8ff6029c8655f638a87bf40772752c328486cd0ed89a931c2613c81fee06e3a717c0f5b00f228d6f1b0192befada0d522fd943da8136f8992b0580aa61173c1867ca623dee558063c965a37",
      "verification":"5521024c7b7fb6c310fccf1ba33b082519d82964ea93868d676662d4a59ad548df0e7d21025bdf3f181f53e9696227843950deb72dcd374ded17c057159513c3d0abe20b6421035e819642a8915a2572f972ddbdbe3042ae6437349295edce9bdc3b8884bbf9a32103b209fd4f53a7170ea4444e0cb0a6bb6a53c2bd016926989cf85f9b0fba17a70c2103b8d9d5771d8f513aa0869b9cc8d50986403b78c6da36890638c3d46a5adce04a2102ca0e27697b9c248f6f16e085fd0061e26f44da85b58ee835c110caa5ec3ba5542102df48f60e8f3e01c48ff40b9b7f1310d7a8b2a193188befe1c2e3df740e89509357ae"
   },
   "size":2064,
   "time":1584568869,
   "tx":[
      {
         "attributes":[

         ],
         "net_fee":"0",
         "nonce":2632263850,
         "scripts":[

         ],
         "size":10,
         "sys_fee":"0",
         "txid":"0x47e026ec2366be9ab834eb262f21311d28bd6cdfc90b3876e4f5c49d510f3c31",
         "type":"MinerTransaction",
         "version":0,
         "vin":[

         ],
         "vout":[

         ]
      },
      {
         "attributes":[
            {
               "data":"4649f940788c16c01e70d39e3613252ef7a58dbc",
               "usage":"Script"
            }
         ],
         "gas":"0",
         "net_fee":"0",
         "script":"0500bea7d01414b197bb42bb224fe63c20bbc93dd1e183ae1cb120144649f940788c16c01e70d39e3613252ef7a58dbc53c1087472616e7366657267a9fe67930cdd180863395620af7a3f0109d2c0b7f166f470c61e6df75140",
         "scripts":[
            {
               "invocation":"40eab6fbbf8839314cd21118329f8bfc8048c99f1d77851e32a2fb6ffcf56c4cffb021d4b0fc0168af13b2a246566fedee0fa387fffcbc0eda2c5abe9b16f7f52b",
               "verification":"2102eb9704f6676a4e332b0a91057ddfe9bba4126b453e7104f9616f1a081e702694ac"
            }
         ],
         "size":228,
         "sys_fee":"0",
         "txid":"0x086b9af57f2c621894327f95c581f4050d3f96ce818d2b13a473e4ea98aae4f7",
         "type":"InvocationTransaction",
         "version":1,
         "vin":[

         ],
         "vout":[

         ]
      },
      {
         "attributes":[

         ],
         "gas":"0",
         "net_fee":"0",
         "script":"203236613166663031323834313564376166323737396438383038623534306238149a9763d8cd9d160af3162f44f50123afe58bca5652c10b7061794f7264657247617367a87c7976662f9b2012c2a5ace6a0dbd6f532d259",
         "scripts":[
            {
               "invocation":"403a16ac944a11ba2bc470257b8f7ad651e72ace4d8755c8b0a4947b288b4fe0e017215ffd97bf000c225fe8432204d196a16ec85daae5c7523d58e8ef2b97ed56",
               "verification":"2103fc8d713440a5febf9dbd6d61b96234453b5c7a75ba5841c5cf04ec4351a9e3d4ac"
            }
         ],
         "size":360,
         "sys_fee":"0",
         "txid":"0x56028d91ec5630151a9725422ce2dd5a42da04a20e4359402cefd868bef681c9",
         "type":"InvocationTransaction",
         "version":1,
         "vin":[
            {
               "txid":"0x23895058c1f0c5614c30460aead0db1749733496ca1bd9e8582c0f300b7ab99e",
               "vout":1
            }
         ],
         "vout":[
            {
               "address":"AX8kHN1rYFdtGj4V1G4Ncqt75MaVpQS4jp",
               "asset":"0x602c79718b16e442de58778e148d0b1084e3b2dffd5de6b7b16cee7969282de7",
               "n":0,
               "value":"0.00001"
            },
            {
               "address":"AVsH4oZrmtGVDC95m4xY5wNtLZChdFzqiC",
               "asset":"0x602c79718b16e442de58778e148d0b1084e3b2dffd5de6b7b16cee7969282de7",
               "n":1,
               "value":"0.18538"
            }
         ]
      },
      {
         "attributes":[

         ],
         "gas":"0",
         "net_fee":"0",
         "script":"203134366538636239363763346365623361623361366534353132336336303739143a36d8cd56c4dfbdf636a892ed786af4e4c695f352c10b7061794f7264657247617367a87c7976662f9b2012c2a5ace6a0dbd6f532d259",
         "scripts":[
            {
               "invocation":"40c870c95a84c5cc846c1ca5b09621d94d4d7343a44a6b3df3eaa2c8ec9a539232532f0f0668efa28669f531680d0419c346805c74e76fd7d7939fc55e9238c824",
               "verification":"21028032387622db11004d901ac1bea972e0ca417e84c590cf031e7eaffed6e8498dac"
            }
         ],
         "size":360,
         "sys_fee":"0",
         "txid":"0xc10f1cca2eec7ee3fde687eee411ad1c92c7bcbdb715c052d71ba1029967142f",
         "type":"InvocationTransaction",
         "version":1,
         "vin":[
            {
               "txid":"0xae7523b5d50450139282a607320c140876cbf56134a7fa552acae3636edfbb34",
               "vout":1
            }
         ],
         "vout":[
            {
               "address":"AX8kHN1rYFdtGj4V1G4Ncqt75MaVpQS4jp",
               "asset":"0x602c79718b16e442de58778e148d0b1084e3b2dffd5de6b7b16cee7969282de7",
               "n":0,
               "value":"0.00001"
            },
            {
               "address":"AM5gYTDHwBqiMXZ8xmY7ciFDUZ1oEBKcN5",
               "asset":"0x602c79718b16e442de58778e148d0b1084e3b2dffd5de6b7b16cee7969282de7",
               "n":1,
               "value":"0.18562"
            }
         ]
      },
      {
         "attributes":[

         ],
         "net_fee":"0",
         "scripts":[
            {
               "invocation":"408e55b007686bf29cc8cdc2f65e7b5b448e2efe5a68e71c6f35f617edce8603fe376fa9d3edb6ed1295866fcc50db9f0ac7c5f14bf4b12fbd308cea4775389630",
               "verification":"21026c69fc74aaf06273d34fdc2d24382140155714f65901b83b34d3c99a7e4eaf85ac"
            }
         ],
         "size":202,
         "sys_fee":"0",
         "txid":"0xc72ba18597b99a76fb15701d7667926846ede86f0125e4b7222b81857d09c6e5",
         "type":"ContractTransaction",
         "version":0,
         "vin":[
            {
               "txid":"0xa9348f870d0a2630c75f47103d068959ea178c4aa2eecc86509f7a1dac6e2534",
               "vout":0
            }
         ],
         "vout":[
            {
               "address":"ASedaViTE9NdgunCjjeaZDufLcPYQEC3vx",
               "asset":"0xc56f33fc6ecfcd0c225c4ab356fee59390af8560be0e930faebe74a6daff7c9b",
               "n":0,
               "value":"52"
            }
         ]
      },
      {
         "attributes":[
            {
               "data":"59a6e5472a2d026c76e3cb76fc5d01aba86728a6",
               "usage":"Script"
            }
         ],
         "gas":"0",
         "net_fee":"0",
         "script":"050056218300144649f940788c16c01e70d39e3613252ef7a58dbc1459a6e5472a2d026c76e3cb76fc5d01aba86728a653c1087472616e73666572679bc8c394217f198608b06106d101dd1cac5a4b31f1666f3e3341286df702",
         "scripts":[
            {
               "invocation":"40ba5421c8bb68385fc7d7a813acf61180fa04a984275a36586024d907f208af1e05af99e0f040c00ddfe26b02d16485ec724064bb03c3af00112f5f13f4bc8b5c",
               "verification":"210366186cebdb804cfd4af553ab5ffceb6acc99326b312f8cb08e949f29134d8afcac"
            }
         ],
         "size":228,
         "sys_fee":"0",
         "txid":"0xf9d35a9c7258efd474d7ae3827a2377e73399ac8b0c67399034f78b329cd95a4",
         "type":"InvocationTransaction",
         "version":1,
         "vin":[

         ],
         "vout":[

         ]
      }
   ],
   "version":0
}
```

## License
Apache License 2.0

The initial version of this package was built by the O3 team.

-----

Built with ❤️ by the team behind [NeoLogin: a simple and easy to integrate wallet provider for Neo dApps](https://neologin.io).
