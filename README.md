<p align="center">
  <img width="90%" alt="morpheusvm" src="assets/logo.jpeg">
</p>
<p align="center">
  OracleVM
</p>
<p align="center">
  <a href="https://github.com/bianyuanop/oraclevm/actions/workflows/oraclevm-unit-tests.yml"><img src="https://github.com/bianyuanop/oraclevm/actions/workflows/oraclevm-unit-tests.yml/badge.svg" /></a>
</p>

[What is a blockchain oracle?](https://en.wikipedia.org/wiki/Blockchain_oracle)

*A blockchain oracle is a third-party service that connects smart contracts with the outside world, primarily to feed information in from the world, but also the reverse. Information from the world encapsulates multiple sources, so that decentralised knowledge is obtained. Oracles provide a way for the decentralized Web3 ecosystem to access existing data sources, legacy systems, and advanced computations. Decentralized oracle networks (DONs) enable the creation of hybrid smart contracts, where on-chain code and off-chain infrastructure are combined to support advanced decentralized applications (dApps) that react to real-world events and interoperate with traditional systems.*

What is the advantage of OracleVM?

The most advantage of OracleVM is about it can deliver any types of off-chain data to be accessible on-chain. By abstracting different data/aggregation methods into a set of universal interfaces, only with those interfaces get implemented for a specified entity, that kind of entity can be served by OracleVM easily. Also the approach is a subnet oracle solution, which means this approach can have advantages in terms of efficiency, scalability, and reducing the risk of network congestion. 

## Workflow

```
                                 +--------------+
                                 | Stock Feeds  |
                                 +--------------+
                                   :                                                     Verify & Storemorpheusvm Txs
                                   :                                                   +--------------------+
                                   v                                                   v                    |
  +----------------------+       +--------------+  UploadEntity(id, type, payload)   +------------------------+
  |         ...          |   ..> |    Feeds     | ---------------------------------> |          Node          |
  +----------------------+       +--------------+                                    +------------------------+
                                   ^                                                   |
                                   :                                                   | Building block
                                   :                                                   |
+ - - - - - - - - - - - - -+                                                           |
' Node:                    '                                                           |
'                          '                                                           v
' +----------------------+ '     +--------------+                                    +------------------------+
' | EntityCollection map | '     | Sports Feeds |                                    |    Node(next round)    |
' +----------------------+ '     +--------------+                                    +------------------------+
' +----------------------+ '
' |     History map      | '
' +----------------------+ '
'                          '
+ - - - - - - - - - - - - -+
```

Data are submitted by user transactions by action `UploadEntity`, in which users can upload specific `id`, `type`, and `payload`. Each entity instance has an unique `id` and `type`, `id ` is used to locate `EntityCollection` and `History`,  where `type` serves for how to parse uploaded `payload` into `entity`.

```

                                                                      Store aggregation result
                                                                       ............................
                                                                       v                          :
     On building block   +------------------+  Aggregates entities   +------------------------------+
    ...................> |    Aggregator    | .....................> |           History            |
                         +------------------+                        +------------------------------+
                           |
                           |
                           |
                         +------------------+                        +------------------------------+
                         | EntityCollection | ---------------------- |         Entity Log           |
                         +------------------+                        +------------------------------+
```

An aggregator is placed in each `EntityCollection`, which is responsible for aggregating entities on building new blocks. After building a new block, the aggregation results will be stored at memory(`History` here) and database.

### On chain query

```
                   WarpMessage with query result
  +--------------------------------------------------------------------+
  v                                                                    |
+----------------+  warp_message:Query(entityIndex, DestinationID)   +---------------+
| Querier Subnet | ------------------------------------------------> | Oracle Subnet |
+----------------+                                                   +---------------+
```

On chain query is done by sending a warp message to call `Query` action. 

[^Warp Message]: `hypersdk` provides support for Avalanche Warp Messaging (AWM) out-of-the-box. AWM enables any Avalanche Subnet to send arbitrary messages to any another Avalanche Subnet in just a few seconds (or less) without relying on a trusted relayer or bridge (just the validators of the Subnet sending the message). You can learn more about AWM and how it works [here](https://docs.google.com/presentation/d/1eV4IGMB7qNV7Fc4hp7NplWxK_1cFycwCMhjrcnsE9mU/edit).

### Off chain query

```
           Aggregation History for that entity
  +--------------------------------------------------+
  v                                                  |
+--------+  rpc_call:History(entityIndex, limit)   +---------------+
| Client | --------------------------------------> | Oracle Subnet |
+--------+                                         +---------------+
```

## Entity & Aggregation Abstraction

Entity 

```go
type Entity interface {
	Publisher() string
	Tick() int64
	Marshal() []byte
}

func Unmarshal(b []bytes) (Entity, error) 
```

Aggregator

```
type EntityAggregator interface {
	Result() (Entity, error)
	MergeOne(Entity)
	RemoveOne(Entity)
}
```

### Interfaces implementation example - Stock prices data

```go
type Stock struct {
	Ticker string `json:"ticker"`
	Price  uint64 `json:"price"`

	publisher crypto.PublicKey
	tick      int64
}
# Publisher: return publisher -> string
# Tick: return tick
# Marshal: json.Marshal(stock)
```

```go
type StockAggregator struct {
	ticker string
	sum    uint64
	count  uint64
}
# MergeOne: sum += e.Price, count++
# Result: return Entity(&Stock{ Price: sum/count, Ticker: ticker})
# RemoveOne: sum -= e.Price, count--
```

## TODOs

+ Test on fuji testnet for wrap message query

+ Serve historical aggregation results for warp queries

+ Find solution to serve non-continuously entities, e.g. sports match result, which usually results in hours rather than seconds(stocks prices)

+ Allow users to submit a tick along with their upload transaction to prevent duplicate transaction & duplication check in one block

+ Implement Staking component & credit component, which provides reputation for aggregation


## Developer Guides

**Since OracleVM is modified based on [morpheusvm](https://github.com/ava-labs/hypersdk/tree/main/examples/morpheusvm), except the guides on how to add a new types of entity, all other guides are exactly the same.** 

The first step to running this demo is to launch your own `oracle` Subnet. You
can do so by running the following command from this location (may take a few
minutes):

```bash
./scripts/run.sh;
```

When the Subnet is running, you'll see the following logs emitted:
```
cluster is ready!
avalanche-network-runner is running in the background...

use the following command to terminate:

./scripts/stop.sh;
```

_By default, this allocates all funds on the network to `morpheus1rvzhmceq997zntgvravfagsks6w0ryud3rylh4cdvayry0dl97nsp30ucp`. The private
key for this address is `0x323b1d8f4eed5f0da9da93071b034f2dce9d2d22692c172f3cb252a64ddfafd01b057de320297c29ad0c1f589ea216869cf1938d88c9fbd70d6748323dbf2fa7`.
For convenience, this key has is also stored at `demo.pk`._

### Build `morpheus-cli`
To make it easy to interact with the `morpheusvm`, we implemented the `morpheus-cli`.
Next, you'll need to build this tool. You can use the following command:

```bash
./scripts/build.sh
```

_This command will put the compiled CLI in `./build/morpheus-cli`._

### Configure `morpheus-cli`
Next, you'll need to add the chains you created and the default key to the
`morpheus-cli`. You can use the following commands from this location to do so:
```bash
./build/morpheus-cli key import demo.pk
```

If the key is added corretcly, you'll see the following log:
```
database: .morpheus-cli
imported address: morpheus1rvzhmceq997zntgvravfagsks6w0ryud3rylh4cdvayry0dl97nsp30ucp
```

Next, you'll need to store the URLs of the nodes running on your Subnet:
```bash
./build/morpheus-cli chain import-anr
```

If `morpheus-cli` is able to connect to ANR, it will emit the following logs:
```
database: .morpheus-cli
stored chainID: 2mQy8Q9Af9dtZvVM8pKsh2rB3cT3QNLjghpet5Mm5db4N7Hwgk uri: http://127.0.0.1:45778/ext/bc/2mQy8Q9Af9dtZvVM8pKsh2rB3cT3QNLjghpet5Mm5db4N7Hwgk
stored chainID: 2mQy8Q9Af9dtZvVM8pKsh2rB3cT3QNLjghpet5Mm5db4N7Hwgk uri: http://127.0.0.1:58191/ext/bc/2mQy8Q9Af9dtZvVM8pKsh2rB3cT3QNLjghpet5Mm5db4N7Hwgk
stored chainID: 2mQy8Q9Af9dtZvVM8pKsh2rB3cT3QNLjghpet5Mm5db4N7Hwgk uri: http://127.0.0.1:16561/ext/bc/2mQy8Q9Af9dtZvVM8pKsh2rB3cT3QNLjghpet5Mm5db4N7Hwgk
stored chainID: 2mQy8Q9Af9dtZvVM8pKsh2rB3cT3QNLjghpet5Mm5db4N7Hwgk uri: http://127.0.0.1:14628/ext/bc/2mQy8Q9Af9dtZvVM8pKsh2rB3cT3QNLjghpet5Mm5db4N7Hwgk
stored chainID: 2mQy8Q9Af9dtZvVM8pKsh2rB3cT3QNLjghpet5Mm5db4N7Hwgk uri: http://127.0.0.1:44160/ext/bc/2mQy8Q9Af9dtZvVM8pKsh2rB3cT3QNLjghpet5Mm5db4N7Hwgk
```

_`./build/morpheus-cli chain import-anr` connects to the Avalanche Network Runner server running in
the background and pulls the URIs of all nodes tracking each chain you
created._


### Check Balance
To confirm you've done everything correctly up to this point, run the
following command to get the current balance of the key you added:
```bash
./build/morpheus-cli key balance
```

If successful, the balance response should look like this:
```
database: .morpheus-cli
address: morpheus1rvzhmceq997zntgvravfagsks6w0ryud3rylh4cdvayry0dl97nsp30ucp
chainID: 2mQy8Q9Af9dtZvVM8pKsh2rB3cT3QNLjghpet5Mm5db4N7Hwgk
uri: http://127.0.0.1:45778/ext/bc/2mQy8Q9Af9dtZvVM8pKsh2rB3cT3QNLjghpet5Mm5db4N7Hwgk
balance: 1000.000000000 RED
```

### Generate Another Address
Now that we have a balance to send, we need to generate another address to send to. Because
we use bech32 addresses, we can't just put a random string of characters as the reciepient
(won't pass checksum test that protects users from sending to off-by-one addresses).
```bash
./build/morpheus-cli key generate
```

If successful, the `morpheus-cli` will emit the new address:
```
database: .morpheus-cli
created address: morpheus1s3ukd2gnhxl96xa5spzg69w7qd2x4ypve0j5vm0qflvlqr4na5zsezaf2f
```

By default, the `morpheus-cli` sets newly generated addresses to be the default. We run
the following command to set it back to `demo.pk`:
```bash
./build/morpheus-cli key set
```

You should see something like this:
```
database: .morpheus-cli
chainID: 2mQy8Q9Af9dtZvVM8pKsh2rB3cT3QNLjghpet5Mm5db4N7Hwgk
stored keys: 2
0) address: morpheus1rvzhmceq997zntgvravfagsks6w0ryud3rylh4cdvayry0dl97nsp30ucp balance: 1000.000000000 RED
1) address: morpheus1s3ukd2gnhxl96xa5spzg69w7qd2x4ypve0j5vm0qflvlqr4na5zsezaf2f balance: 0.000000000 RED
set default key: 0
```

### Send Tokens
Lastly, we trigger the transfer:
```bash
./build/morpheus-cli action transfer
```

The `morpheus-cli` will emit the following logs when the transfer is successful:
```
database: .morpheus-cli
address: morpheus1rvzhmceq997zntgvravfagsks6w0ryud3rylh4cdvayry0dl97nsp30ucp
chainID: 2mQy8Q9Af9dtZvVM8pKsh2rB3cT3QNLjghpet5Mm5db4N7Hwgk
balance: 1000.000000000 RED
recipient: morpheus1s3ukd2gnhxl96xa5spzg69w7qd2x4ypve0j5vm0qflvlqr4na5zsezaf2f
✔ amount: 10
continue (y/n): y
✅ txID: sceRdaoqu2AAyLdHCdQkENZaXngGjRoc8nFdGyG8D9pCbTjbk
```

### Bonus: Watch Activity in Real-Time
To provide a better sense of what is actually happening on-chain, the
`morpheus-cli` comes bundled with a simple explorer that logs all blocks/txs that
occur on-chain. You can run this utility by running the following command from
this location:
```bash
./build/morpheus-cli chain watch
```

If you run it correctly, you'll see the following input (will run until the
network shuts down or you exit):
```
database: .morpheus-cli
available chains: 1 excluded: []
0) chainID: 2mQy8Q9Af9dtZvVM8pKsh2rB3cT3QNLjghpet5Mm5db4N7Hwgk
select chainID: 0
uri: http://127.0.0.1:45778/ext/bc/2mQy8Q9Af9dtZvVM8pKsh2rB3cT3QNLjghpet5Mm5db4N7Hwgk
watching for new blocks on 2mQy8Q9Af9dtZvVM8pKsh2rB3cT3QNLjghpet5Mm5db4N7Hwgk 👀
height:1 txs:1 units:440 root:WspVPrHNAwBcJRJPVwt7TW6WT4E74dN8DuD3WXueQTMt5FDdi
✅ sceRdaoqu2AAyLdHCdQkENZaXngGjRoc8nFdGyG8D9pCbTjbk actor: morpheus1rvzhmceq997zntgvravfagsks6w0ryud3rylh4cdvayry0dl97nsp30ucp units: 440 summary (*actions.Transfer): [10.000000000 RED -> morpheus1s3ukd2gnhxl96xa5spzg69w7qd2x4ypve0j5vm0qflvlqr4na5zsezaf2f]
```

<br>
<br>
<br>

<p align="center">
  <a href="https://github.com/ava-labs/hypersdk"><img width="40%" alt="powered-by-hypersdk" src="assets/hypersdk.png"></a>
</p>
