package protocol

import "fmt"

func SendToGateway(sensorID, payload string) {
	fmt.Printf("[Websocket] Sensor %s sending: %s\n", sensorID, payload)
}
