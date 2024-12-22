// Package clientbridge allows for multiple different modbus servers to be bridged together over different protocols.
package clientbridge

import (
	"context"

	"go.viam.com/rdk/components/board"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"

	"viam-modbus/common"
)

type resourceType = sensor.Sensor
type config = *modbusBridgeConfig

var API = board.API
var Model = resource.NewModel("viam-soleng", "modbus", "client-bridge")

func init() {
	resource.RegisterComponent(
		API,
		Model,
		resource.Registration[resourceType, config]{
			Constructor: NewModbusBridge,
		},
	)
}

func NewModbusBridge(ctx context.Context, deps resource.Dependencies, conf resource.Config, logger logging.Logger) (resourceType, error) {
	logger.Infof("Starting Modbus Board Component %v", common.Version)
	c, cancelFunc := context.WithCancel(context.Background())
	b := modbusBridge{
		Named:      conf.ResourceName().AsNamed(),
		logger:     logger,
		cancelFunc: cancelFunc,
		ctx:        c,
	}

	if err := b.Reconfigure(ctx, deps, conf); err != nil {
		logger.Errorf("Failed to start Modbus Board Component %v", err)
		return nil, err
	}
	logger.Info("Modbus Board Component started successfully")
	return &b, nil
}

type modbusBridge struct {
	resource.Named
	logger     logging.Logger
	cancelFunc context.CancelFunc
	ctx        context.Context
}

func (b *modbusBridge) Reconfigure(ctx context.Context, deps resource.Dependencies, conf resource.Config) error {
	return nil
}

func (b *modbusBridge) Close(ctx context.Context) error {
	return nil
}

func (b *modbusBridge) Readings(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return nil, nil
}
