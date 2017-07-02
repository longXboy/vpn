package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/songgao/water"
)

func main() {
	var conn net.Conn
	var isClient bool
	var err error
	mode := flag.String("mode", "server", "client or server")

	serverAddr := flag.String("server", "192.168.0.105:5050", "address of server")
	flag.Parse()

	config := water.Config{
		DeviceType: water.TUN,
	}
	if mode != nil {
		if *mode == "client" {
			isClient = true
		}
	}
	if isClient {
		config.Name = "tun3"
	} else {
		config.Name = "tun2"
	}
	config.Persist = true
	ifce, err := water.New(config)
	if err != nil {
		panic(err)
	}

	if isClient {
		conn, err = net.Dial("tcp", *serverAddr)
		if err != nil {
			panic(err)
		}
	} else {
		lis, err := net.Listen("tcp", "0.0.0.0:5050")
		if err != nil {
			panic(err)
		}
		fmt.Println("listen")
		conn, err = lis.Accept()
		if err != nil {
			panic(err)
		}
	}
	go readFromConn(conn, ifce)

	for {
		var b []byte = make([]byte, 1500)
		n, err := ifce.Read(b)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("read from ifce:", n)
		packetData := b[:n]
		b2 := make([]byte, 2)
		binary.BigEndian.PutUint16(b2, uint16(n))

		_, err = conn.Write(b2)
		if err != nil {
			panic(err)
		}
		_, err = conn.Write(packetData)
		if err != nil {
			panic(err)
		}
	}
}

func readFromConn(conn net.Conn, ifce *water.Interface) {
	for {
		var numBytes []byte = make([]byte, 2)
		_, err := conn.Read(numBytes)
		if err != nil {
			panic(err)
		}
		var b []byte = make([]byte, binary.BigEndian.Uint16(numBytes))
		_, err = io.ReadFull(conn, b)
		if err != nil {
			panic(err)
		}
		fmt.Println("read from conn:", binary.BigEndian.Uint16(numBytes))
		_, err = ifce.Write(b)
		if err != nil {
			panic(err)
		}
	}
}
