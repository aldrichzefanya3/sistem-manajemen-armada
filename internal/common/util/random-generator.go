package util

import (
	"fmt"
	"math/rand"
)

var numberPool = []int{1234, 2345, 3456, 4567, 5678, 6789}

func RandomVehicleID() string {
	letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	num := numberPool[rand.Intn(len(numberPool))]

	return fmt.Sprintf(
		"B%d%c%c%c",
		num,
		letters[rand.Intn(len(letters))],
		letters[rand.Intn(len(letters))],
		letters[rand.Intn(len(letters))],
	)
}

func RandomLatitude() float64 {
	return rand.Float64()*180 - 90
}

func RandomLongitude() float64 {
	return rand.Float64()*360 - 180
}
