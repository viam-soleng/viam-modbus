package main

import (
	moduleutils "github.com/thegreatco/viamutils/module"
	"go.viam.com/rdk/module"
	viamutils "go.viam.com/utils"

	modbus_board "viam-modbus/board"
	common "viam-modbus/common"
	modbus_sensor "viam-modbus/sensor"
	"viam-modbus/serverbridge"
)

func main() {
	moduleutils.AddModularResource(modbus_sensor.API, modbus_sensor.Model)
	moduleutils.AddModularResource(modbus_board.API, modbus_board.Model)
	moduleutils.AddModularResource(serverbridge.API, serverbridge.Model)
	viamutils.ContextualMain(moduleutils.RunModule, module.NewLoggerFromArgs(common.LoggerName))
}
