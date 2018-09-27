package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/qshuai/go-electrum/electrum"
)

func main() {
	electrum.DebugMode = true

	node := electrum.NewNode()

	go func() {
		for {
			select {
			case <-node.Error:
				log.Println("node shutdown")
			default:

			}
		}
	}()

	// The hard code for connection server as for can not reference to the
	// variable `serverAddr` via the command `go run keepalive.go`. You
	// should to update it if not available.
	if err := node.ConnectTCP("39.104.125.149:9629"); err != nil {
		log.Fatal(err)
	}

	// sync the process
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		var times int

		for {
			if err := node.Ping(); err != nil {
				log.Println(err)
				wg.Done()
			}

			times++
			fmt.Printf("ping %d times;\n", times)

			time.Sleep(5 * time.Second)
		}
	}()

	wg.Wait()
}
