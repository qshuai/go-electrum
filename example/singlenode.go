package main

import (
	"log"
	"time"

	"github.com/d4l3k/go-electrum/electrum"
)

var (
	// Should specify a available server(IP:PORT)
	serverAddr = "39.104.125.149:9629"

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
	log.Printf("Version: %s", version)

	banner, err := node.ServerBanner()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Banner: %s", banner)

	address, err := node.ServerDonationAddress()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Address: %s", address)

	peers, err := node.ServerPeersSubscribe()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Peers: %+v", peers)

	headerChan, err := node.BlockchainHeadersSubscribe()
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		for header := range headerChan {
			log.Printf("Headers: %+v", header)
		}
	}()

	hashChan, err := node.BlockchainAddressSubscribe(bitcoinAddress)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		for hash := range hashChan {
			log.Printf("Address history hash: %+v", hash)
		}
	}()

	history, err := node.BlockchainAddressGetHistory(bitcoinAddress)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Address history: %+v", history)

	transaction, err := node.BlockchainTransactionGet("3b885123e87a6f7dbaf1e3bd9e4bf63f1c6d09a6e00ac651596ba56f4d99e85c")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Transaction: %s", transaction)

	transactions, err := node.BlockchainAddressListUnspent(bitcoinAddress)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Unspent transactions: %+v", transactions)

	// TODO(d4l3k) seems to not work, need to subscribe first maybe?
	balance, err := node.BlockchainAddressGetBalance(bitcoinAddress)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Address balance: %+v", balance)

	// now you can deposit some coins to the bitcoinAddress,
	// or deposit/withdraw some coins to your specified address.
	// waite a moment and you will get notification from server
	// about balance and transaction.
	time.Sleep(15 * time.Minute)
}
