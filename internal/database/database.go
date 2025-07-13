package protocol

import "fmt"

func SendToGateway(sensorID, payload string) {
	fmt.Printf("[Database] Sensor %s sending: %s\n", sensorID, payload)
}
