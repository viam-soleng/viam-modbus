package modbus_sensor

import (
	"context"
	"encoding/json"
	"sync"

	"go.viam.com/rdk/components/generic"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"

	"viam-modbus/utils"
)

var Model = resource.NewModel("viam-soleng", "sensor", "modbus-tcp")

// Registers the sensor model
func init() {
	resource.RegisterComponent(
		sensor.API,
		Model,
		resource.Registration[sensor.Sensor, *ModbusSensorConfig]{
			Constructor: NewModbusSensor,
		},
	)
}

// Creates a new modbus sensor instance
func NewModbusSensor(ctx context.Context, deps resource.Dependencies, conf resource.Config, logger logging.Logger) (sensor.Sensor, error) {
	logger.Infof("Starting Modbus Sensor Component %v", utils.Version)
	c, cancelFunc := context.WithCancel(context.Background())
	b := ModbusSensor{
		Named:      conf.ResourceName().AsNamed(),
		logger:     logger,
		cancelFunc: cancelFunc,
		ctx:        c,
	}

	if err := b.Reconfigure(ctx, deps, conf); err != nil {
		return nil, err
	}
	return &b, nil
}

type ModbusSensor struct {
	resource.Named
	mu                sync.RWMutex
	logger            logging.Logger
	cancelFunc        context.CancelFunc
	ctx               context.Context
	blocks            []ModbusBlocks
	modbus_connection resource.Named
	component_type    string
	component_desc    string
}

// Validate ensures all parts of the config are valid and important fields exist.
// Returns implicit dependencies based on the config.
// The path is the JSON path in your robot's config (not the `Config` struct) to the
// resource being validated; e.g. "components.0".
// func (r *ModbusSensor) Validate(path string) ([]string, []string, error) {
// 	// Add config validation code here
// 	var reqDeps []string
//   if r.ModbusConnectionName == "" {
//     return nil, nil, resource.NewConfigValidationFieldRequiredError(path, "modbus_connection_name")
//   }
//   reqDeps = append(reqDeps, r.ModbusConnectionName)
//   return reqDeps, nil, nil
// }

// Returns modbus register values
func (r *ModbusSensor) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// stringify the r.blocks array of json objects to pass it as an interface to the modbus_connection DoCommand()
	// ints become float64 when passing a json object through a interface
	jsonBytes, err := json.Marshal(r.blocks)
	if err != nil {
		r.logger.Infof("Error marshaling JSON:%v", err)
		return nil, err
	}
	jsonBlocks := string(jsonBytes)

	// Create an empty map to store the register key/values
	modbusResponse := make(map[string]interface{})
	//r.logger.Infof("ModbusSensor Readings() calling ModbusConnection DoCommand() with %v", jsonBlocks)
	modbusResponse, err = r.modbus_connection.DoCommand(ctx, map[string]interface{}{"blocks": jsonBlocks})

	// Add the opinionated component key/value attributes to the response
	if r.component_type != "" {
		modbusResponse["component_type"] = r.component_type
	}
	if r.component_desc != "" {
		modbusResponse["component_description"] = r.component_desc
	}

	// Debug - print the map[]interface{}
	//r.logger.Info(modbusResponse)

	// Return the modbus Response and any component attributes
	return modbusResponse, err
}

// Closes the modbus sensor instance
func (r *ModbusSensor) Close(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logger.Info("Closing Modbus Sensor Component")
	r.cancelFunc()

	return nil
}

// DoCommand currently not implemented
func (*ModbusSensor) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{"ok": 1}, nil
}

// Configures the modbus sensor instance
func (r *ModbusSensor) Reconfigure(ctx context.Context, deps resource.Dependencies, conf resource.Config) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logger.Debug("Reconfiguring Modbus Sensor Component")

	newConf, err := resource.NativeConfig[*ModbusSensorConfig](conf)
	if err != nil {
		return err
	}

	// In case the module has changed name
	r.Named = conf.ResourceName().AsNamed()

	return r.reconfigure(newConf, deps)
}

// reconfigures the modbus sensor instance
func (r *ModbusSensor) reconfigure(newConf *ModbusSensorConfig, deps resource.Dependencies) error {
	modbus_connection, _ := deps.Lookup(generic.Named(string(newConf.ModbusConnection)))
	r.modbus_connection = modbus_connection.(resource.Named)
	r.component_type = string(newConf.ComponentType)
	r.component_desc = string(newConf.ComponentDesc)

	r.blocks = newConf.Blocks
	return nil
}
