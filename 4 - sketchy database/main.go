package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

type operation string

const (
	opRetrieve operation = "retrieve"
	opStore    operation = "store"
	opVersion  operation = "version"
)

type database struct {
	store map[string]string
	mu    sync.Mutex
}

func (d *database) startDatabase() {
	d.store = make(map[string]string)
}

func (d *database) retrieveKey(key string) (v string) {
	d.mu.Lock()
	retrievedKey := d.store[key]
	d.mu.Unlock()
	return retrievedKey
}

func (d *database) storeKey(key string, value string) {
	if key == "version" {
		return
	}
	d.mu.Lock()
	d.store[key] = value
	d.mu.Unlock()
}

func parsePacketOperation(packet string) operation {
	if packet == "version" {
		return opVersion
	}

	if !strings.Contains(packet, "=") {
		return opRetrieve
	}

	return opStore
}

func splitStorePacket(packet string) (k, v string) {
	split := strings.SplitN(packet, "=", 2)
	return split[0], split[1]
}

func handleRequest(request string, conn net.PacketConn, address net.Addr) {

	log.Println("Got request:", request)

	op := parsePacketOperation(request)

	switch op {
	case opRetrieve:
		v := DB.retrieveKey(request)
		conn.WriteTo([]byte(fmt.Sprintf("%s=%s", request, v)), address)
	case opStore:
		k, v := splitStorePacket(request)
		DB.storeKey(k, v)
	case opVersion:
		version := "version=Alex super sick database; v1.0.0"
		conn.WriteTo([]byte(version), address)
	}

	return

}

var DB database

func main() {
	DB.startDatabase()

	conn, err := net.ListenPacket("udp", ":9999")
	defer conn.Close()
	log.Println("listening on port :9999")
	if err != nil {
		log.Fatal(err)
	}

	buffer := make([]byte, 1024)

	for {
		n, addr, err := conn.ReadFrom(buffer)

		if n > 1000 {
			continue
		}

		fmt.Println("before goroutine:", string(buffer[:n]))

		handleRequest(string(buffer[:n]), conn, addr)

		if err != nil {
			log.Print(err)
		}
	}

}
