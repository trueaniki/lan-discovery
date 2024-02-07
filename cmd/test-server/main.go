package main

import (
	"fmt"

	"github.com/trueaniki/lan-discovery/discovery"
)

func main() {
	conn, err := discovery.ListenForDiscover(nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Write
	fmt.Printf("Connection: %v\n", conn)
	conn.Write([]byte("HEOEW!!!!!"))

	// Read
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(buf[:n]))
}
