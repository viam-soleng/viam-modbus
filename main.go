package main

import (
	"context"

	"go.viam.com/rdk/components/board"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/module"
	"go.viam.com/utils"

	"viam-modbus/modbus_tcp"
	"viam-modbus/modbus_tcp_board"
	module_utils "viam-modbus/utils"
)

func main() {
	utils.ContextualMain(mainWithArgs, module.NewLoggerFromArgs(module_utils.LoggerName))
}

func mainWithArgs(ctx context.Context, args []string, logger logging.Logger) (err error) {
	custom_module, err := module.NewModuleFromArgs(ctx, logger)
	if err != nil {
		return err
	}

	err = custom_module.AddModelFromRegistry(ctx, sensor.API, modbus_tcp.Model)
	if err != nil {
		return err
	}

	err = custom_module.AddModelFromRegistry(ctx, board.API, modbus_tcp_board.Model)
	if err != nil {
		return err
	}

	err = custom_module.Start(ctx)
	defer custom_module.Close(ctx)
	if err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}
