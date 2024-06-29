package udp

import (
	"log"
	"net"
)

const (
	UDP_PORT = 12345
	BUF_SIZE = 1024
)

func StartLogServer() {
	addr := net.UDPAddr{
		Port: UDP_PORT,
		IP:   net.ParseIP("0.0.0.0"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Fatal("Error starting UDP server: %v\n", err)
	}
	defer conn.Close()

	log.Printf("UDP server listening on %s:%d\n", addr.IP.String(), addr.Port)

	buffer := make([]byte, BUF_SIZE)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Error reading from UDP: %v\n", err)
			continue
		}
		message := string(buffer[:n])
		log.Printf("Received message from %s: %s\n", remoteAddr.String(), message)
	}
}
