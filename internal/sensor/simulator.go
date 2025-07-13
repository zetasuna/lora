package sensor

// import (
// 	"fmt"
// 	"math/rand/v2"
//
// 	"github.com/brianvoe/gofakeit/v7"
// )
//
// type Sensor struct {
// 	ID       string
// 	Type     string
// 	Interval int
// }
//
// func NewSensor(id, sensorType string, interval int) *Sensor {
// 	return &Sensor{
// 		ID:       id,
// 		Type:     sensorType,
// 		Interval: interval,
// 	}
// }
//
// func (s *Sensor) GenerateData() string {
// 	gofakeit.Seed(100000)
// 	switch s.Type {
// 	case "temperature":
// 		return fmt.Sprintf("%.2f°C", 20+rand.Float64()*10)
// 	case "humidity":
// 		return fmt.Sprintf("%.2f%%", 40+rand.Float64()*20)
// 	case "pressure":
// 		return fmt.Sprintf("%.2fhPa", 1000+rand.Float64()*50)
// 	default:
// 		return "N/A"
// 	}
// }
