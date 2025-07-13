package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	lorawan "doan/internal/lora"

	lorapkg "github.com/brocaar/lorawan"
)

type UplinkPayload struct {
	Data string `json:"data"`
}

var sensorKeys = map[string]lorawan.LoRaWANContext{
	"01000001": {
		DevAddr: lorapkg.DevAddr{0x01, 0x00, 0x00, 0x01},
		AppSKey: lorapkg.AES128Key{0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F},
		NwkSKey: lorapkg.AES128Key{0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2A, 0x2B, 0x2C, 0x2D, 0x2E, 0x2F},
	},
	"01000002": {
		DevAddr: lorapkg.DevAddr{0x01, 0x00, 0x00, 0x02},
		AppSKey: lorapkg.AES128Key{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F, 0x20},
		NwkSKey: lorapkg.AES128Key{0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2A, 0x2B, 0x2C, 0x2D, 0x2E, 0x2F, 0x30},
	},
	"01000003": {
		DevAddr: lorapkg.DevAddr{0x01, 0x00, 0x00, 0x03},
		AppSKey: lorapkg.AES128Key{0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F, 0x20, 0x21},
		NwkSKey: lorapkg.AES128Key{0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2A, 0x2B, 0x2C, 0x2D, 0x2E, 0x2F, 0x30, 0x31},
	},
}

func main() {
	http.HandleFunc("/uplink", handleUplink)
	log.Println("Network Server listening on http://localhost:10000/uplink")
	log.Fatal(http.ListenAndServe(":10000", nil))
}

func handleUplink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var payload UplinkPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	binData, err := base64.StdEncoding.DecodeString(payload.Data)
	if err != nil {
		http.Error(w, "Invalid base64", http.StatusBadRequest)
		return
	}

	// extract DevAddr for context lookup
	var phy lorapkg.PHYPayload
	if err := phy.UnmarshalBinary(binData); err != nil {
		http.Error(w, "Invalid lorapkg packet", http.StatusBadRequest)
		return
	}

	macPL, ok := phy.MACPayload.(*lorapkg.MACPayload)
	if !ok {
		http.Error(w, "Invalid MACPayload", http.StatusBadRequest)
		return
	}

	devAddr := macPL.FHDR.DevAddr
	ctx, ok := sensorKeys[devAddr.String()]
	if !ok {
		http.Error(w, "Unknown DevAddr", http.StatusUnauthorized)
		return
	}
	ctx.FCnt = macPL.FHDR.FCnt
	ctx.FPort = *macPL.FPort

	decrypted := lorawan.Decode(ctx, binData)
	fmt.Printf("[NET SERVER] DevAddr=%s FCnt=%d Payload=%s\n", devAddr, ctx.FCnt, string(decrypted))

	// forward to App Server
	resp, err := http.Post("http://localhost:9999/sensor", "application/json", bytes.NewBuffer(decrypted))
	if err != nil {
		log.Println("[NET SERVER] Failed to forward to App Server:", err)
		return
	}
	defer resp.Body.Close()
	w.WriteHeader(http.StatusOK)
}
