package main

import (
	"fmt"
	"net"
	"time"
)

func stream() {
	fmt.Println("Hello, playground")

	// connect to site
	conn, err := net.Dial("tcp", "192.168.7.78:30003")
	if err != nil {
		fmt.Printf("failed to connect: %s\n", err)
		return
	}

	fmt.Printf("connected\n")

	conn.SetDeadline(time.Now().Add(time.Second * 5))

	// Send our HTTP1.1 GET /
	nwrite, err := conn.Write([]byte("GET  HTTP/1.1\n\n"))
	if err != nil {
		fmt.Printf("failed to write request: %s\n", err)
	}

	fmt.Printf("write %d bytes\n", nwrite)
	// read data
	// create our buffer
	buffer := make([]byte, 2048)
	nread, err := conn.Read(buffer)
	if err != nil {
		fmt.Printf("failed to read from socket: %s\n", err)
		return
	}
	fmt.Printf("bytes read: %d, content: %s\n", nread, buffer)

}
