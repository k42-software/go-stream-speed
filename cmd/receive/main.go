// receive a stream over the network
package main

import (
	"log"
	"sync"
)

//
const serverHost = "127.0.0.1"
const kcpPort = "7951"
const tcpPort = "7961"
const quicPort = "7971"
const tlsPort = "7991"

func main() {
	log.SetFlags(log.Ltime)

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go listenTcp(wg)

	wg.Add(1)
	go listenTls(wg)

	wg.Add(1)
	go listenQuic(wg)

	wg.Add(1)
	go listenKcp(wg)

	wg.Wait()
}
