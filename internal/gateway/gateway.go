package gateway

import (
	"bytes"
	"io"
	"log"
	"net/http"

	lorawan "doan/internal/lora"
)

type Gateway struct {
	ListenAddr string
	ForwardURL string
	DecodeLoRa bool
}

// Constructor
func NewGateway(addr, forward string, decode bool) *Gateway {
	return &Gateway{
		ListenAddr: addr,
		ForwardURL: forward,
		DecodeLoRa: decode,
	}
}

// Start HTTP server
func (g *Gateway) Start() error {
	http.HandleFunc("/uplink", g.handleUplink)
	log.Println("[Gateway] Listening on", g.ListenAddr)
	return http.ListenAndServe(g.ListenAddr, nil)
}

// Xử lý POST /uplink
func (g *Gateway) handleUplink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var decoded []byte
	if g.DecodeLoRa {
		decoded = lorawan.LorawanDecode(body)
	} else {
		decoded = body
	}

	log.Println("[Gateway] Received LoRa data, decoded:", string(decoded))

	go g.forward(decoded)
	w.WriteHeader(http.StatusOK)
}

// Gửi dữ liệu tới network server
func (g *Gateway) forward(data []byte) {
	resp, err := http.Post(g.ForwardURL+"/ingest", "application/json", bytes.NewReader(data))
	if err != nil {
		log.Println("[Gateway] Forward error:", err)
		return
	}
	defer resp.Body.Close()

	log.Println("[Gateway] Forwarded to", g.ForwardURL, "Status:", resp.Status)
}
