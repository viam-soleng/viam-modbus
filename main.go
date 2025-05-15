package main

import (
	"go.viam.com/rdk/components/board"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/module"
	"go.viam.com/rdk/resource"
	//viamutils "go.viam.com/utils"

	modbus_board  "viam-modbus/modbus_board"
	modbus_sensor "viam-modbus/modbus_sensor"
	//module_utils  "viam-modbus/utils"
)

func main() {
	// ModularMain can take multiple APIModel arguments, if your module implements multiple models.
	module.ModularMain( 
		resource.APIModel{board.API,  modbus_board.Model},
		resource.APIModel{sensor.API, modbus_sensor.Model},
	)
}
