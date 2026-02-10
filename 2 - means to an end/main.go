package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

type dataBaseEntry struct {
	date  int32
	money int32
}

type duplicateTimestamp struct {
	Msg string
}

func (e *duplicateTimestamp) Error() string {
	return fmt.Sprintf("error: duplicate timestamp")
}

var (
	mu          sync.Mutex
	dataBaseMap map[string][]dataBaseEntry
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	allData := make([]byte, 9)

	for {
		//fmt.Println("looped")
		_, err := io.ReadFull(conn, allData)
		//fmt.Println("Read data")

		if err == io.EOF {
			return
		}

		if err != nil {
			log.Printf("Connection error: %v", err)
		}

		operation := string(allData[0])

		switch operation {
		case "I":
			err := handleRequestAdd(allData, conn)
			if err != nil {
				fmt.Println("error:", err)
				return
			}
			//fmt.Println(conn.RemoteAddr().String(), "wrote to the database")
		case "Q":
			response := handleRequestRead(allData, conn)
			//fmt.Println(conn.RemoteAddr().String(), "read from the database")
			//fmt.Println("response:", response)
			b := make([]byte, 4)
			binary.BigEndian.PutUint32(b, uint32(response))
			conn.Write(b)

		default:
			return
		}

	}

}

func handleRequestRead(data []byte, conn net.Conn) int32 {

	minTime := int32(binary.BigEndian.Uint32(data[1:5]))
	maxTime := int32(binary.BigEndian.Uint32(data[5:9]))
	var total int64
	var count int64

	if minTime > maxTime {
		fmt.Printf("mintime %v greater than maxtime %v\n", minTime, maxTime)
		return 0
	}

	mu.Lock()
	dataBase := dataBaseMap[conn.RemoteAddr().String()]
	for _, entry := range dataBase {
		if entry.date >= minTime && entry.date <= maxTime {
			count++
			total += int64(entry.money)
			if count <= 5 || count > int64(len(dataBase))-5 {
				fmt.Printf("  Match #%d: date=%d, money=%d\n", count, entry.date, entry.money)
			}
		}
	}
	mu.Unlock()
	if count == 0 {
		return 0
	}
	result := int32(total / count)
	fmt.Printf("Query [%d, %d]: count=%d, total=%d, avg=%d, dbSize=%d\n",
		minTime, maxTime, count, total, result, len(dataBase))

	return result
}

func handleRequestAdd(data []byte, conn net.Conn) error {
	timestamp := int32(binary.BigEndian.Uint32(data[1:5]))
	value := int32(binary.BigEndian.Uint32(data[5:9]))
	newDatabaseObject := dataBaseEntry{
		date:  timestamp,
		money: value,
	}

	fmt.Printf("Insert: timestamp=%d, value=%d\n", timestamp, value)

	mu.Lock()
	defer mu.Unlock()

	dataBase := dataBaseMap[conn.RemoteAddr().String()]

	if len(dataBase) == 0 {
		dataBase = append(dataBase, dataBaseEntry{date: timestamp, money: value})
		dataBaseMap[conn.RemoteAddr().String()] = dataBase
		return nil
	}
	for idx, i := range dataBase {
		if timestamp == i.date {
			err := fmt.Errorf("duplicate dates detected")
			return err
		}
		if timestamp > i.date {
			if (idx + 1) == len(dataBase) {
				dataBase = append(dataBase, newDatabaseObject)
				dataBaseMap[conn.RemoteAddr().String()] = dataBase
				return nil
			}

		} else {
			dataBase = append(dataBase, dataBaseEntry{})
			copy(dataBase[idx+1:], dataBase[idx:])
			dataBase[idx] = newDatabaseObject
			dataBaseMap[conn.RemoteAddr().String()] = dataBase
			return nil
		}
	}

	return nil
}

func main() {
	dataBaseMap = make(map[string][]dataBaseEntry)

	// Listen for incoming connections on port 8080
	ln, err := net.Listen("tcp", ":9999")
	defer ln.Close()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Listening on port 9999")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("error accepting connection: %v", err)
			continue
		}
		go handleConnection(conn)

	}
}
