package main

import (
	"go.viam.com/rdk/components/board"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/module"
	"go.viam.com/rdk/resource"

	modbus_board "viam-modbus/modbus_board"
	modbus_sensor "viam-modbus/modbus_sensor"
	module_utils "viam-modbus/utils"
)

func main() {
	module.ModularMain(module_utils.LoggerName,
	                   resource.APIModel{sensor.API, modbus_sensor.Model},
	                   resource.APIModel{board.API, modbus_board.Model},
	                   )
}
