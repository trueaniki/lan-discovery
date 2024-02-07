package main

import "github.com/trueaniki/lan-discovery/discovery"

func main() {
	conn, err := discovery.RunDiscovery()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Read
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		panic(err)
	}
	println(string(buf[:n]))

	// Write
	conn.Write([]byte("HEOEW!!!!!"))
}
