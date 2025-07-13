package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"time"
)

type UplinkPayload struct {
	Data string `json:"data"`
}

func main() {
	listenAddr, err := net.ResolveUDPAddr("udp", ":10001")
	if err != nil {
		log.Fatalf("failed to resolve UDP addr: %v", err)
	}

	conn, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		log.Fatalf("failed to listen on UDP: %v", err)
	}
	defer conn.Close()
	log.Println("Gateway listening on 127.0.0.1:10001 (UDP)")

	for {
		buf := make([]byte, 2048)
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("read error: %v", err)
			continue
		}

		log.Printf("Received %d bytes from sensor %s", n, addr.String())

		// Encode as base64 for HTTP
		b64 := base64.StdEncoding.EncodeToString(buf[:n])
		payload := UplinkPayload{Data: b64}
		jsonBytes, err := json.Marshal(payload)
		if err != nil {
			log.Printf("json marshal error: %v", err)
			continue
		}

		// Send HTTP POST to network server
		req, err := http.NewRequest("POST", "http://localhost:10000/uplink", bytes.NewReader(jsonBytes))
		if err != nil {
			log.Printf("http request error: %v", err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("http post error: %v", err)
			continue
		}
		resp.Body.Close()

		log.Printf("Forwarded packet to net server via HTTP POST")
	}
}
