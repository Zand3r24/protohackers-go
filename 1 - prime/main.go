package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
)

type request struct {
	Method string `json:"method"`
	Number int    `json:"number"`
}

type response struct {
	Method string `json:"method"`
	Prime  bool   `json:"prime"`
}

func main() {
	ln, err := net.Listen("tcp", "127.0.0.1:9999")
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go connectionHandler(conn)
	}
}

func isPrime(num int) bool {
	if (num % 2) == 0 {
		return false
	}
	for i := 3; i < num-1; i += 2 {
		if num%i == 0 {
			return false
		}
	}
	return true
}

func sendMalformed(conn net.Conn) error {
	response, err := json.Marshal(response{Method: "fuck"})
	if err != nil {
		return err
	}

	_, err = conn.Write(response)
	if err != nil {
		return err
	}

	return nil
}

func sendCorrect(conn net.Conn, prime bool) error {
	response, err := json.Marshal(response{Method: "isPrime", Prime: prime})
	if err != nil {
		return err
	}

	_, err = conn.Write(response)
	if err != nil {
		return err
	}

	return nil

}

func connectionHandler(conn net.Conn) {
	defer conn.Close()

	buff := bufio.NewReader(conn)
	var primeRequest request
	for {
		req, err := buff.ReadBytes('\n')
		if err != nil {

			if err == io.EOF {
				fmt.Println("The client has closed the connection")
				return
			}

			sendMalformed(conn)
			fmt.Println(err)
			continue
		}

		err = json.Unmarshal(req, &primeRequest)
		if err != nil {
			sendMalformed(conn)
			fmt.Println(err)
			continue
		}

		if primeRequest.Method != "isPrime" {
			sendMalformed(conn)
			fmt.Println("Method was incorrect, got", primeRequest.Method, ". Expected isPrime")
			continue
		}

		prime := isPrime(primeRequest.Number)

		err = sendCorrect(conn, prime)
		if err != nil {
			fmt.Println(err)
			continue
		}

	}

}
