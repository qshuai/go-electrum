package main

import (
	"fmt"
	"log"
	"time"

	"github.com/qshuai/go-electrum/electrum"
)

var (
	// Should specify a available server(IP:PORT) if connecting to the
	// following server failed.
	serverAddr = "39.104.125.149:9629"

	// A bitcoin testnet address which should be consistent with the server
	// network type
	bitcoinAddress = "n4FyJMDYXJmPEm7cffFLrwLXvGWn8cW9q2"
)

func main() {
	node := electrum.NewNode()
	if err := node.ConnectTCP(serverAddr); err != nil {
		log.Fatal(err)
	}

	version, err := node.ServerVersion()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Version: %v\n\n", version)

	banner, err := node.ServerBanner()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Banner: %s\n\n", banner)

	address, err := node.ServerDonationAddress()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Address: %s\n\n", address)

	peers, err := node.ServerPeersSubscribe()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Peers: %+v\n\n", peers)

	headerChan, err := node.BlockchainHeadersSubscribe()
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		for header := range headerChan {
			fmt.Printf("Headers: %+v\n\n", header)
		}
	}()

	hashChan, err := node.BlockchainAddressSubscribe(bitcoinAddress)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		for hash := range hashChan {
			fmt.Printf("Address history hash: %+v\n\n", hash)
		}
	}()

	history, err := node.BlockchainAddressGetHistory(bitcoinAddress)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Address history: %+v\n\n", history)

	transaction, err := node.BlockchainTransactionGet(
		"3b885123e87a6f7dbaf1e3bd9e4bf63f1c6d09a6e00ac651596ba56f4d99e85c", true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Transaction: %+v\n\n", transaction)

	transactions, err := node.BlockchainAddressListUnspent(bitcoinAddress)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Unspent transactions: %+v\n\n", transactions)

	balance, err := node.BlockchainAddressGetBalance(bitcoinAddress)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Address balance: %+v\n\n", balance)

	// Send server.ping request in order to keep alive connection to
	// electrum server
	go func() {
		for {
			if err := node.Ping(); err != nil {
				log.Fatal(err)
			}

			time.Sleep(5 * time.Second)
		}
	}()

	// Now you can deposit some coins to the bitcoinAddress,
	// or deposit/withdraw some coins to your specified address.
	// waite a moment and you will get notification from server
	// about balance and transaction.
	time.Sleep(15 * time.Minute)
}
