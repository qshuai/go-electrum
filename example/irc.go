package main

import (
	"log"

	"github.com/qshuai/go-electrum/irc"
)

func main() {
	log.Println(irc.FindElectrumServers())
}
