package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"regexp"
	"strings"
)

const CLOSE_CONNECTION string = "CLOSE_CONNECTION"

type proxyConnection struct {
	upstream          net.Conn
	downstream        net.Conn
	upstreamChannel   chan string
	downstreamChannel chan string
}

func (p *proxyConnection) initProxy() {
	p.connectUpstream()
	upstreamChannel := make(chan string)
	p.upstreamChannel = upstreamChannel
	downstreamChannel := make(chan string)
	p.downstreamChannel = downstreamChannel
}

func (p *proxyConnection) forwardMessageUpstream(str string) {
	replacedString := replaceCryptoAddress(str)
	log.Println("upstream message:", replacedString)
	io.WriteString(p.upstream, replacedString)
}

func (p *proxyConnection) forwardMessageDownstream(str string) {
	replacedString := replaceCryptoAddress(str)
	log.Println("downstream message:", replacedString)
	io.WriteString(p.downstream, replacedString)
}

func (p *proxyConnection) readMessageDownstream() {
	reader := bufio.NewReader(p.downstream)
	for {
		msg, err := reader.ReadString('\n')
		log.Println("reading downstream message:", msg)
		if err != nil {
			log.Println("downstream error:", err)
			p.upstreamChannel <- CLOSE_CONNECTION
			return
		}

		p.upstreamChannel <- msg
	}
}

func (p *proxyConnection) readMessageUpstream() {
	reader := bufio.NewReader(p.upstream)
	for {
		msg, err := reader.ReadString('\n')
		log.Println("reading upstream message:", msg)
		if err != nil {
			log.Println("upstream error:", err)
			p.downstreamChannel <- CLOSE_CONNECTION
			return
		}

		p.downstreamChannel <- msg
	}
}

func (p *proxyConnection) connectUpstream() {
	upstream, err := net.Dial("tcp", "chat.protohackers.com:16963")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to upstream")
	p.upstream = upstream
}

func (p *proxyConnection) closeConnections() {
	p.downstream.Close()
	p.upstream.Close()
}

func replaceCryptoAddress(str string) string {
	re := regexp.MustCompile(`7[a-zA-Z0-9]{25,34}`)

	// Find all matches with their positions
	matches := re.FindAllStringIndex(str, -1)
	if matches == nil {
		return str
	}

	// Build result by replacing valid addresses
	var result strings.Builder
	lastIndex := 0

	for _, match := range matches {
		start, end := match[0], match[1]

		// Check character before
		validBefore := start == 0 || !isAlphaNumOrHyphen(str[start-1])

		// Check character after
		validAfter := end >= len(str) || !isAlphaNumOrHyphen(str[end])

		// Write everything up to this match
		result.WriteString(str[lastIndex:start])

		if validBefore && validAfter {
			// Valid address - replace it
			result.WriteString("7YWHMfk9JZe0LM0g1ZauHuiSxhI")
		} else {
			// Not valid - keep original
			result.WriteString(str[start:end])
		}

		lastIndex = end
	}

	// Write remaining string
	result.WriteString(str[lastIndex:])

	return result.String()
}

func isAlphaNumOrHyphen(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-'
}

func handleProxy(client net.Conn) {
	var proxy proxyConnection
	proxy.initProxy()

	proxy.downstream = client
	defer proxy.closeConnections()

	go proxy.readMessageDownstream()
	go proxy.readMessageUpstream()

	for {
		select {
		case data := <-proxy.downstreamChannel:
			log.Println("downstream channel:", data)
			if data == CLOSE_CONNECTION {
				return
			}
			proxy.forwardMessageDownstream(data)
		case data := <-proxy.upstreamChannel:
			log.Println("upstream channel:", data)
			if data == CLOSE_CONNECTION {
				return
			}
			proxy.forwardMessageUpstream(data)
		}
	}

}

func main() {
	log.Println("Starting server...")
	ln, err := net.Listen("tcp", ":9999")
	log.Println("Listening on port :9999")
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := ln.Accept()

		if err != nil {
			log.Println(err)
			continue
		}

		go handleProxy(conn)
	}

}
