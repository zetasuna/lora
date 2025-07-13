package protocol

import "fmt"

func SendToGateway(sensorID, payload string) {
	fmt.Printf("[MQTT] Sensor %s sending: %s\n", sensorID, payload)
}
