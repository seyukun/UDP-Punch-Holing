package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

type clientType map[string]bool

var clients = clientType{}

func (c clientType) keys(filter string) string {
	output := []string{}
	for key := range c {
		if key != filter {
			output = append(output, key)
		}
	}

	return strings.Join(output, ",")
}

func Server() {
	localAddress := ":50000"
	if len(os.Args) > 2 {
		localAddress = os.Args[2]
	}

	addr, _ := net.ResolveUDPAddr("udp", localAddress)
	conn, _ := net.ListenUDP("udp", addr)

	for {
		buffer := make([]byte, 1024)

		if n, remoteAddr, err := conn.ReadFromUDP(buffer); err != nil {
			fmt.Printf("[ERROR] %s\n", err)
		} else {
			message := string(buffer[:n])
			fmt.Println("[INCOMING]", message)
			if len(message) > 8 && message[:8] != "register" {
				continue
			}

			clients[remoteAddr.String()] = true

			for client := range clients {
				resp := clients.keys(client)
				if len(resp) > 0 {
					r, _ := net.ResolveUDPAddr("udp", client)
					conn.WriteTo([]byte(resp), r)
					fmt.Printf("[INFO] Responded to %s with %s\n", client, string(resp))
				}
			}
		}
	}
}
