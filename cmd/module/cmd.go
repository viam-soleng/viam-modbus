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
	)
}

/*
package main

import (
	"go.viam.com/rdk/module"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/components/board"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/components/generic"

	modbus_board      "viam-modbus/modbus_board"
	modbus_sensor     "viam-modbus/modbus_sensor"
	modbus_connection "viam-modbus/modbus_connection"
)

func main() {
	// ModularMain can take multiple APIModel arguments, if your module implements multiple models.
	module.ModularMain(
		resource.APIModel{board.API,   modbus_board.Model},
		resource.APIModel{sensor.API,  modbus_sensor.Model},
		resource.APIModel{generic.API, modbus_connection.Model},
	)
}
*/
