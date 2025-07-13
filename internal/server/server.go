package server

import (
	"fmt"

	"github.com/brocaar/lorawan"
)

type Gateway struct {
	ID string
}

func NewGateway(id string) *Gateway {
	return &Gateway{
		ID: id,
	}
}

func (g *Gateway) LorawanDecode() {
	nwkSKey := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	appSKey := [16]byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

	var phy lorawan.PHYPayload
	// use use UnmarshalBinary when decoding a byte-slice
	if err := phy.UnmarshalText([]byte("gAQDAgEDAAAGcwcK4mTU9+EX0sA=")); err != nil {
		panic(err)
	}

	ok, err := phy.ValidateUplinkDataMIC(lorawan.LoRaWAN1_0, 0, 0, 0, nwkSKey, lorawan.AES128Key{})
	if err != nil {
		panic(err)
	}
	if !ok {
		panic("invalid mic")
	}

	if err := phy.DecodeFOptsToMACCommands(); err != nil {
		panic(err)
	}

	phyJSON, err := phy.MarshalJSON()
	if err != nil {
		panic(err)
	}

	if err := phy.DecryptFRMPayload(appSKey); err != nil {
		panic(err)
	}
	macPL, ok := phy.MACPayload.(*lorawan.MACPayload)
	if !ok {
		panic("*MACPayload expected")
	}

	pl, ok := macPL.FRMPayload[0].(*lorawan.DataPayload)
	if !ok {
		panic("*DataPayload expected")
	}

	fmt.Println(string(phyJSON))
	fmt.Println(string(pl.Bytes))
}

func (g *Gateway) LorawanJoinRequestRead() {
	var phy lorawan.PHYPayload
	if err := phy.UnmarshalText([]byte("AAQDAgEEAwIBBQQDAgUEAwItEGqZDhI=")); err != nil {
		panic(err)
	}

	jrPL, ok := phy.MACPayload.(*lorawan.JoinRequestPayload)
	if !ok {
		panic("MACPayload must be a *JoinRequestPayload")
	}

	fmt.Println(phy.MHDR.MType)
	fmt.Println(jrPL.JoinEUI)
	fmt.Println(jrPL.DevEUI)
	fmt.Println(jrPL.DevNonce)
}

func (g *Gateway) LorawanJoinAcceptSend() {
	appKey := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	joinEUI := lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1}
	devNonce := lorawan.DevNonce(258)

	phy := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.JoinAccept,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.JoinAcceptPayload{
			JoinNonce:  65793,
			HomeNetID:  [3]byte{2, 2, 2},
			DevAddr:    lorawan.DevAddr([4]byte{1, 2, 3, 4}),
			DLSettings: lorawan.DLSettings{RX2DataRate: 0, RX1DROffset: 0},
			RXDelay:    0,
		},
	}

	// set the MIC before encryption
	if err := phy.SetDownlinkJoinMIC(lorawan.JoinRequestType, joinEUI, devNonce, appKey); err != nil {
		panic(err)
	}
	if err := phy.EncryptJoinAcceptPayload(appKey); err != nil {
		panic(err)
	}

	str, err := phy.MarshalText()
	if err != nil {
		panic(err)
	}

	bytes, err := phy.MarshalBinary()
	if err != nil {
		panic(err)
	}

	fmt.Println(string(str))
	fmt.Println(string(bytes))
}
