package main

import (
	"fmt"
	"io"
	"net"
)

func main() {
	ln, err := net.Listen("tcp", "192.168.1.55:8080")

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Listening on 192.168.1.55:8080")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println("Connection Accepted")
		go connectionHandler(conn)

	}
}

func connectionHandler(conn net.Conn) {
	defer conn.Close()
	data := make([]byte, 1024)
	var buffer []byte
	var totalBytesRead int

	for {
		bytesRead, err := conn.Read(data)
		totalBytesRead += bytesRead
		if err != nil {
			if err == io.EOF {
				buffer = append(buffer, data[:bytesRead]...)
				fmt.Println("Read a total of", totalBytesRead, "bytes")
				break
			}
			fmt.Println(err)
			return
		}
		buffer = append(buffer, data[:bytesRead]...)
	}

	fmt.Println("Finished reading data")
	fmt.Printf("%#v", buffer)
	bytesWritten, err := conn.Write(buffer[:totalBytesRead])

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(bytesWritten)

}
