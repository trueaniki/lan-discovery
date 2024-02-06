package main

import (
	"fmt"
	"lan-discovery/discovery"
)

func main() {
	conn, err := discovery.Discover()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Read
	fmt.Printf("Connection: %v\n", conn)
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(buf[:n]))

	// Write
	conn.Write([]byte("HEOEW!!!!!"))
}
