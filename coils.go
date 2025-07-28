package viammodbus

import (
	"context"
	"fmt"

	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
)

var CoilSensorModel = NamespaceFamily.WithModel("coils")

func init() {
	resource.RegisterComponent(
		sensor.API,
		CoilSensorModel,
		resource.Registration[sensor.Sensor, *coilSensorConfig]{
			Constructor: newCoilSensor,
		})
}

// TODO: Add the ability to read multiple coils at once
type coilSensorConfig struct {
	ModbusClient string `json:"modbus_client"`
	Offset       uint16 `json:"offset"`
	UnitID       uint8  `json:"unit_id"`
}

func (cfg *coilSensorConfig) Validate(path string) ([]string, []string, error) {
	if cfg.ModbusClient == "" {
		return nil, nil, fmt.Errorf("modbus_client is required")
	}
	return []string{cfg.ModbusClient}, nil, nil
}

type coilSensor struct {
	resource.AlwaysRebuild
	name   resource.Name
	logger logging.Logger
	config *coilSensorConfig
	client *modbusClient
}

func newCoilSensor(ctx context.Context, deps resource.Dependencies, config resource.Config, logger logging.Logger) (sensor.Sensor, error) {
	newConf, err := resource.NativeConfig[*coilSensorConfig](config)
	if err != nil {
		return nil, err
	}

	cs := &coilSensor{
		name:   config.ResourceName(),
		logger: logger,
		config: newConf,
	}

	// Get the modbus client from the global registry
	client, err := GlobalClientRegistry.Get(newConf.ModbusClient)
	if err != nil {
		return nil, err
	}
	cs.client = client

	return cs, nil
}

func (cs *coilSensor) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("Coil component completed yet")
	/*
		value, err := cs.client.ReadCoil(cs.config.Offset, cs.config.UnitID)
		if err != nil {
			cs.logger.Errorf("Failed to read coil at offset %d: %v", cs.config.Offset, err)
			return nil, err
		}

		readings := map[string]interface{}{
			"coil": value,
		}

		return readings, nil
	*/
}

func (cs *coilSensor) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("DoCommand not implemented")
}

func (cs *coilSensor) Close(ctx context.Context) error {
	return nil
}

func (cs *coilSensor) Name() resource.Name {
	return cs.name
}
