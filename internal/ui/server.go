package protocol

import "fmt"

func SendToGateway(sensorID, payload string) {
	fmt.Printf("[Web] Sensor %s sending: %s\n", sensorID, payload)
}
