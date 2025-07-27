package main

import (
	viammodbus "github.com/viam-soleng/viam-modbus"

	"go.viam.com/rdk/components/generic"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/module"
	"go.viam.com/rdk/resource"
)

func main() {
	module.ModularMain(
		resource.APIModel{API: generic.API, Model: viammodbus.ModbusClientModel},
		resource.APIModel{API: sensor.API, Model: viammodbus.ModbusSensorModel},
		resource.APIModel{API: sensor.API, Model: viammodbus.CoilSensorModel},
	)
}
