package main

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"regexp"
	"slices"
	"strings"
	"sync"
)

type chatUser struct {
	userName         string
	connection       net.Conn
	connectionString string
}

type broker struct {
	mu         sync.Mutex
	clients    []*chatUser
	msgChannel chan messageType
}

type messageType struct {
	connectionString string
	msg              string
}

func (b *broker) addClient(client *chatUser) {

	userJoinedMessage := fmt.Sprintf("* user %s has joined", client.userName)
	logger.Info(userJoinedMessage)
	joinMessage := messageType{
		connectionString: client.connection.RemoteAddr().String(),
		msg:              userJoinedMessage,
	}
	b.sendMessage(joinMessage)
	b.mu.Lock()
	b.clients = append(b.clients, client)
	b.mu.Unlock()

}

func (b *broker) removeClient(client *chatUser) {
	b.mu.Lock()
	fmt.Printf("%#v\n\n", b.clients)
	for i, v := range b.clients {
		if client.connectionString == v.connectionString {
			b.clients = slices.Delete(b.clients, i, i+1)
			break
		}

	}
	b.mu.Unlock()
	userDisconnectMessage := fmt.Sprintf("* user %s has left like a bitch", client.userName)
	logger.Info(userDisconnectMessage)
	disconnectMessage := messageType{
		connectionString: client.connectionString,
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

func (b *broker) listUsers(user *chatUser) string {
	b.mu.Lock()
	var connectedClients string
	for idx, client := range b.clients {
		if client.connection.RemoteAddr().String() == user.connection.RemoteAddr().String() {
			continue
		}
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
		logger.Error(err.Error())
		return
	}

	name = strings.TrimSpace(name)

	if strings.Contains(name, " ") {
		conn.Close()
		return
	}

	if len(name) > 20 {
		conn.Close()
		return
	}

	if !regexp.MustCompile(`^[a-zA-Z0-9]*$`).MatchString(name) {
		conn.Close()
		return
	}

	client := chatUser{
		userName:         name,
		connection:       conn,
		connectionString: conn.RemoteAddr().String(),
	}

	connectedClientStrings := fmt.Sprintf("* The room contains: %s\n",
		messageBroker.listUsers(&client))

	logger.Info("client connected", "name", name, "room", connectedClientStrings)

	io.WriteString(conn, connectedClientStrings)
	logger.Info(fmt.Sprintf("Announced room members to %s", name))

	messageBroker.addClient(&client)
	defer messageBroker.removeClient(&client)

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
var logger *slog.Logger

func main() {
	logOpts := &slog.HandlerOptions{AddSource: false}
	logger = slog.New(slog.NewTextHandler(os.Stdout, logOpts))

	messageBroker.msgChannel = make(chan messageType)
	go messageBroker.startBroker()

	ln, err := net.Listen("tcp", ":9999")
	if err != nil {
		logger.Error("couldnt start server")
		os.Exit(2)
	}

	fmt.Println("server started")

	for {
		connection, err := ln.Accept()
		if err != nil {
			logger.Error("Error accepting connection", "connection", connection.RemoteAddr().String())
			continue
		}
		logger.Info("Connection accepted:", "conn:", connection.RemoteAddr().String())
		go clientConnect(connection)
	}
}
