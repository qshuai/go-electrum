# go-electrum [![GoDoc](https://godoc.org/github.com/qshuai/go-electrum?status.svg)](https://godoc.org/github.com/qshuai/go-electrum)
This repository is a fork of [d4l3k/go-electrum](https://github.com/d4l3k/go-electrum) that is unmaintained now. 

A pure Go [Electrum](https://electrum.org/) bitcoin library. This makes it easy to write bitcoin based services using Go without having to run a full bitcoin node.

![go-electrum](https://rawgit.com/qshuai/go-electrum/master/media/logo.png)

This is very much WIP and has a number of unimplemented methods. This will eventually be rewritten into a more Go-esque library and handle wallet generation.

Packages provided

* [electrum](https://godoc.org/github.com/qshuai/go-electrum/electrum) - Library for using JSON-RPC to talk directly to Electrum servers.
* [wallet](https://godoc.org/github.com/qshuai/go-electrum/wallet) - A bitcoin wallet built on [btcwallet](https://github.com/btcsuite/btcwallet) with Electrum as the backend. **Notice: not available at the current vertsion**.
* [irc](https://godoc.org/github.com/qshuai/go-electrum/irc) - A helper module for finding electrum servers using the [#electrum IRC channel](http://docs.electrum.org/en/latest/protocol.html?highlight=irc#server-peers-subscribe) on Freenode. **Notice: not supported**.

## Usage
See [example/](https://github.com/qshuai/go-electrum/tree/master/example) for more.

### electrum [![GoDoc](https://godoc.org/github.com/qshuai/go-electrum/electrum?status.svg)](https://godoc.org/github.com/qshuai/go-electrum/electrum)
```bash
$ go get -u github.com/qshuai/go-electrum/electrum
```

```go
package main

import (
  "log"

  "github.com/qshuai/go-electrum/electrum"
)

func main() {
	node := electrum.NewNode()
    // the specified ip is testnet server
	if err := node.ConnectTCP("39.104.125.149:9629"); err != nil {
		log.Fatal(err)
	}
    
    // please use bitcoin address accordant to the server environment
	balance, err := node.BlockchainAddressGetBalance("n4FyJMDYXJmPEm7cffFLrwLXvGWn8cW9q2")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Address balance: %+v", balance)
}
```

# License
go-electrum is licensed under the MIT license.

Made by [Tristan Rice](https://fn.lc).
