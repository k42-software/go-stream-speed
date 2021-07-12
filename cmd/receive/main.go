// receive a stream over the network
package main

import (
	"log"
	"sync"
)

//
const serverHost = "127.0.0.1"
const udpTcpPort = "7931"
const relyPort = "7941"
const kcpPort = "7951"
const tcpPort = "7961"
const quicPort = "7971"
const tlsPort = "7991"

func main() {
	log.SetFlags(log.Ltime)

	log.Println("[DEBUG] Starting")

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go listenTcp(wg)

	wg.Add(1)
	go listenTls(wg)

	wg.Add(1)
	go listenQuic(wg)

	wg.Add(1)
	go listenKcp(wg)

	wg.Add(1)
	go listenUdpTcp(wg)

	wg.Wait()
}
