package main

import (
	moduleutils "github.com/thegreatco/viamutils/module"
	"go.viam.com/rdk/module"
	viamutils "go.viam.com/utils"

	"viam-modbus/board"
	"viam-modbus/clientbridge"
	common "viam-modbus/common"
	"viam-modbus/sensor"
	"viam-modbus/server"
	"viam-modbus/serverbridge"
)

func main() {
	moduleutils.AddModularResource(sensor.API, sensor.Model)
	moduleutils.AddModularResource(board.API, board.Model)
	moduleutils.AddModularResource(serverbridge.API, serverbridge.Model)
	moduleutils.AddModularResource(server.API, server.Model)
	moduleutils.AddModularResource(clientbridge.API, clientbridge.Model)
	viamutils.ContextualMain(moduleutils.RunModule, module.NewLoggerFromArgs(common.LoggerName))
}
