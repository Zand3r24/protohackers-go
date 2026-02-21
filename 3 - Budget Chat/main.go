package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"slices"
	"sync"
)

type chatUser struct {
	userName   string
	connection net.Conn
}

type broker struct {
	mu         sync.Mutex
	clients    []chatUser
	msgChannel chan messageType
}

type messageType struct {
	connectionString string
	msg              string
}

func (b *broker) addClient(client chatUser) {
	b.mu.Lock()
	b.clients = append(b.clients, client)
	b.mu.Unlock()
	userJoinedMessage := fmt.Sprintf("* user %s has joined", client.userName)
	joinMessage := messageType{
		connectionString: "",
		msg:              userJoinedMessage,
	}
	b.sendMessage(joinMessage)
}

func (b *broker) removeClient(client chatUser) {
	b.mu.Lock()
	for i, v := range b.clients {
		if client.connection.RemoteAddr().String() == v.connection.RemoteAddr().String() {
			client.connection.Close()
			b.clients = slices.Delete(b.clients, i, i+1)
		}

	}
	b.mu.Unlock()
	userDisconnectMessage := fmt.Sprintf("* user %s has left like a bitch", client.userName)
	disconnectMessage := messageType{
		connectionString: "",
		msg:              userDisconnectMessage,
	}
	b.sendMessage(disconnectMessage)
}

func (b *broker) sendMessage(message messageType) {
	b.msgChannel <- message
}

func (b *broker) startBroker() {
	for message := range b.msgChannel {
		b.mu.Lock()
		for _, chatClient := range b.clients {
			if chatClient.connection.RemoteAddr().String() == message.connectionString {
				continue
			}
			io.WriteString(chatClient.connection, message.msg+"\n")
		}
		b.mu.Unlock()
	}
}

func (b *broker) listUsers() string {
	b.mu.Lock()
	var connectedClients string
	for idx, client := range b.clients {
		connectedClients += client.userName
		if idx != len(messageBroker.clients)-1 {
			connectedClients += ", "
		}
	}
	b.mu.Unlock()

	return connectedClients
}

func clientConnect(conn net.Conn) {
	io.WriteString(conn, "What's yer fuckin name?\n")
	reader := bufio.NewReader(conn)
	name, err := reader.ReadString('\n')

	if err != nil {
		io.WriteString(conn, "Shits fucked\n")
		return
	}

	connectedClientStrings := fmt.Sprintf("* The room contains: %s\n", messageBroker.listUsers())

	io.WriteString(conn, connectedClientStrings)

	client := chatUser{
		userName:   name,
		connection: conn,
	}

	messageBroker.addClient(client)
	defer messageBroker.removeClient(client)

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		message := scanner.Text()
		messageFormat := fmt.Sprintf("[%s] %s", name, message)
		userMessage := messageType{
			connectionString: conn.RemoteAddr().String(),
			msg:              messageFormat,
		}
		messageBroker.sendMessage(userMessage)
	}

}

var messageBroker broker

func main() {
	go messageBroker.startBroker()
	ln, err := net.Listen("tcp", ":9999")
	if err != nil {
		log.Fatal("couldnt start server")
	}

	fmt.Println("server started")

	for {
		connection, err := ln.Accept()
		if err != nil {
			log.Printf("error accepting connection: %v", err)
			continue
		}
		go clientConnect(connection)
	}
}
