package main

import (
	"io"
	"net"
	"slices"
	"sync"
)

type chatUser struct {
	username   string
	connection *net.TCPConn
}

type broker struct {
	mu         sync.Mutex
	clients    []chatUser
	msgChannel chan string
}

func (b *broker) addClient(client chatUser) {
	b.mu.Lock()
	b.clients = append(b.clients, client)
	b.mu.Unlock()
}

func (b *broker) removeClient(client chatUser) {
	b.mu.Lock()
	for i, v := range b.clients {
		if client.connection.RemoteAddr().String() == v.connection.RemoteAddr().String() {
			b.clients = slices.Delete(b.clients, i, i+1)
		}

	}
	b.mu.Unlock()
}

func (b *broker) sendMessage(message string) {
	b.msgChannel <- message
}

func (b *broker) startBroker() {
	for msg := range b.msgChannel {
		b.mu.Lock()
		for _, chatClient := range b.clients {
			io.WriteString(chatClient.connection, chatClient.username+": "+msg)
		}
	}
}
