package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"time"

	m "github.com/seyukun/gomacros"
)

var gSessionId string
var gRegistryUrl string
var gTransactionId []byte
var gPeers []net.UDPAddr

func Client() {
	sessionId, err := makeSessionId(24)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	} else {
		gSessionId = sessionId
	}
	ready()
}

func ready() {
	registryUrl := os.Args[1]
	if registryUrl == "" {
		fmt.Fprintln(os.Stderr, "registry url is required")
		os.Exit(1)
	} else {
		gRegistryUrl = registryUrl
	}

	localAddress := ":51820"
	if len(os.Args) > 2 {
		localAddress = os.Args[2]
	}

	local, _ := net.ResolveUDPAddr("udp", localAddress)
	conn, _ := net.ListenUDP("udp", local)

	handler(conn)
}

func handler(conn *net.UDPConn) {
	// stun
	go func() {
		stun, err := net.ResolveUDPAddr("udp", "stun.l.google.com:19302")
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR][%s:%d] %v\n", m.M__FILE__(), m.M__LINE__(), err)
		}
		for {
			if req, transactionId, err := createStunBindingRequest(); err != nil {
				fmt.Fprintf(os.Stderr, "[ERROR][%s:%d] %v\n", m.M__FILE__(), m.M__LINE__(), err)
			} else {
				gTransactionId = transactionId
				_, err := conn.WriteTo(req, stun)
				if err != nil {
					fmt.Fprintf(os.Stderr, "[ERROR][%s:%d] %v\n", m.M__FILE__(), m.M__LINE__(), err)
				}
			}
			gPeers = getPeers(gRegistryUrl)
			time.Sleep(5 * time.Second)
		}
	}()

	go func() {
		fmt.Println("listening")
		for {
			buffer := make([]byte, 1500)
			go func(size int, addr *net.UDPAddr, err error) {
				if err != nil {
					fmt.Fprintf(os.Stderr, "[ERROR][%s:%d] %v\n", m.M__FILE__(), m.M__LINE__(), err)
				} else {
					// stun response must least 32 bytes (20 + 12)
					if buffer[0]>>6 == 0 &&
						uint32(buffer[4])<<24 == 0x21000000 &&
						uint32(buffer[5])<<16 == 0x00120000 &&
						uint32(buffer[6])<<8 == 0x0000A400 &&
						uint32(buffer[7]) == 0x00000042 &&
						len(buffer) > 32 {
						ip, port, err := parseStunBindingResponse(buffer, gTransactionId)
						if err != nil {
							fmt.Fprintf(os.Stderr, "[Error][%s:%d] %v\n", m.M__FILE__(), m.M__LINE__(), err)
						} else {
							register(gRegistryUrl, fmt.Sprintf("%s:%d", ip.String(), port))
						}
					} else {
						fmt.Printf("[Message] [%s:%d] %s\n", addr.IP.String(), addr.Port, string(buffer[0:size]))
					}
				}
			}(conn.ReadFromUDP(buffer))
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "exit" {
			break
		} else {
			for n := range gPeers {
				go conn.WriteToUDP([]byte(line), &gPeers[n])
			}
		}
	}
}

type RegisterRequest struct {
	Address   string `json:"address"`
	SessionId string `json:"sessionId"`
}

type PeerAddressListResponse struct {
	PeerAddresses []string `json:"peers"`
}

func register(registryUrl string, MyAddress string) {
	body := RegisterRequest{
		Address:   MyAddress,
		SessionId: gSessionId,
	}
	bodyJson, err := json.Marshal(body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Error][%s:%d] %v\n", m.M__FILE__(), m.M__LINE__(), err)
	}
	if resp, err := http.Post(registryUrl, "application/json", bytes.NewReader(bodyJson)); err != nil {
		fmt.Fprintf(os.Stderr, "[Error][%s:%d] %v\n", m.M__FILE__(), m.M__LINE__(), err)
	} else {
		defer resp.Body.Close()
		fmt.Println("[Register] ", MyAddress)
	}
}

func getPeers(registryUrl string) []net.UDPAddr {
	if resp, err := http.Get(registryUrl); err != nil {
		fmt.Fprintf(os.Stderr, "[Error][%s:%d] %v\n", m.M__FILE__(), m.M__LINE__(), err)
		return nil
	} else {
		defer resp.Body.Close()
		var body PeerAddressListResponse
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			fmt.Fprintf(os.Stderr, "[Error][%s:%d] %v\n", m.M__FILE__(), m.M__LINE__(), err)
		}
		var peers []net.UDPAddr
		for n := range body.PeerAddresses {
			if peerAddress, err := net.ResolveUDPAddr("udp", body.PeerAddresses[n]); err != nil {
				fmt.Fprintf(os.Stderr, "[Error][%s:%d] %v\n", m.M__FILE__(), m.M__LINE__(), err)
			} else {
				peers = append(peers, *peerAddress)
			}
		}
		return peers
	}
}

func makeSessionId(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-+=!@#$%^&*()_"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}
