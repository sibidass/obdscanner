package main

import (
	"fmt"

	"github.com/sibidass/obdscanner/internal/device"
)

const (
	ipAddress string = "192.168.0.10"
	port      string = "35000"
)

func main() {
	var dev *device.WifiDevice
	var err error

	devAddress := ipAddress + ":" + port

	dev, err = device.NewWifiDevice(devAddress)

	if err != nil {
		return
	}

	fmt.Printf("Device version: %s\n", dev.getVersion())
}
