package main

import (
	moduleutils "github.com/thegreatco/viamutils/module"
	"go.viam.com/rdk/components/board"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/module"
	viamutils "go.viam.com/utils"

	modbus_board "viam-modbus/modbus_board"
	modbus_sensor "viam-modbus/modbus_sensor"
	"viam-modbus/utils"
)

func main() {
	moduleutils.AddModularResource(sensor.API, modbus_sensor.Model)
	moduleutils.AddModularResource(board.API, modbus_board.Model)
	viamutils.ContextualMain(moduleutils.RunModule, module.NewLoggerFromArgs(utils.LoggerName))
}
