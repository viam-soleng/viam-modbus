package main

import (
	"context"

	"go.viam.com/rdk/components/board"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/module"
	"go.viam.com/utils"

	modbus_board "viam-modbus/modbus_board"
	modbus_sensor "viam-modbus/modbus_sensor"
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

	err = custom_module.AddModelFromRegistry(ctx, sensor.API, modbus_sensor.Model)
	if err != nil {
		return err
	}

	err = custom_module.AddModelFromRegistry(ctx, board.API, modbus_board.Model)
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
