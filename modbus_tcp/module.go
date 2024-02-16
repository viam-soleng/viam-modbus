package modbus_tcp

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/simonvetter/modbus"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"

	"viam-modbus/utils"
)

var Model = resource.NewModel("viam-soleng", "comm", "modbus-tcp")

func init() {
	resource.RegisterComponent(
		sensor.API,
		Model,
		resource.Registration[sensor.Sensor, *ModbusTcpConfig]{
			Constructor: NewModbusTcpSensor,
		},
	)
}

func NewModbusTcpSensor(ctx context.Context, deps resource.Dependencies, conf resource.Config, logger logging.Logger) (sensor.Sensor, error) {
	logger.Infof("Starting Modbus TCP Sensor Component %v", utils.Version)
	c, cancelFunc := context.WithCancel(context.Background())
	b := ModbusTcpSensor{
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

type ModbusTcpSensor struct {
	resource.Named
	mu         sync.RWMutex
	logger     logging.Logger
	cancelFunc context.CancelFunc
	ctx        context.Context
	client     *modbus.ModbusClient
	uri        string
	timeout    time.Duration
	blocks     []ModbusBlocks
}

// Readings implements sensor.Sensor.
func (r *ModbusTcpSensor) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.client == nil {
		return nil, errors.New("modbus client not initialized")
	}
	results := map[string]interface{}{}
	for _, block := range r.blocks {
		switch block.Type {
		case "coils":
			b, err := r.ReadCoils(uint16(block.Offset), uint16(block.Length))
			if err != nil {
				return nil, err
			}
			writeBoolArrayToOutput(b, block, results)
		case "discrete_inputs":
			b, err := r.ReadDiscreteInputs(uint16(block.Offset), uint16(block.Length))
			if err != nil {
				return nil, err
			}
			writeBoolArrayToOutput(b, block, results)
		case "holding_registers":
			b, err := r.ReadHoldingRegisters(uint16(block.Offset), uint16(block.Length))
			if err != nil {
				return nil, err
			}
			writeUInt16ArrayToOutput(b, block, results)
		case "input_registers":
			b, err := r.ReadInputRegisters(uint16(block.Offset), uint16(block.Length))
			if err != nil {
				return nil, err
			}
			writeUInt16ArrayToOutput(b, block, results)
		}
	}
	return results, nil
}

func writeBoolArrayToOutput(b []bool, block ModbusBlocks, results map[string]interface{}) {
	for i, v := range b {
		field_name := block.Name + "." + fmt.Sprint(i)
		results[field_name] = v
	}
}

func writeUInt16ArrayToOutput(b []uint16, block ModbusBlocks, results map[string]interface{}) {
	for i, v := range b {
		field_name := block.Name + "." + fmt.Sprint(i)
		results[field_name] = v
	}
}

func (r *ModbusTcpSensor) ReadCoils(offset, length uint16) ([]bool, error) {
read:
	b, err := r.client.ReadCoils(uint16(offset), uint16(length))
	if err != nil {
		r.logger.Warn("Failed to read modbus client, got EOF, reinitializing modbus client and retrying...")
		err := r.initializeModbusClient()
		if err != nil {
			return nil, err
		}
		goto read
	}
	return b, nil
}

func (r *ModbusTcpSensor) ReadDiscreteInputs(offset, length uint16) ([]bool, error) {
read:
	b, err := r.client.ReadDiscreteInputs(uint16(offset), uint16(length))
	if err != nil {
		r.logger.Warn("Failed to read modbus client, got EOF, reinitializing modbus client and retrying...")
		err := r.initializeModbusClient()
		if err != nil {
			return nil, err
		}
		goto read
	}
	return b, nil
}

func (r *ModbusTcpSensor) ReadHoldingRegisters(offset, length uint16) ([]uint16, error) {
read:
	b, err := r.client.ReadRegisters(uint16(offset), uint16(length), modbus.HOLDING_REGISTER)
	if err != nil {
		r.logger.Warn("Failed to read modbus client, got EOF, reinitializing modbus client and retrying...")
		err := r.initializeModbusClient()
		if err != nil {
			return nil, err
		}
		goto read
	}
	return b, nil
}

func (r *ModbusTcpSensor) ReadInputRegisters(offset, length uint16) ([]uint16, error) {
read:
	b, err := r.client.ReadRegisters(uint16(offset), uint16(length), modbus.INPUT_REGISTER)
	if err != nil {
		r.logger.Warn("Failed to read modbus client, got EOF, reinitializing modbus client and retrying...")
		err := r.initializeModbusClient()
		if err != nil {
			return nil, err
		}
		goto read
	}
	return b, nil
}

// Close implements resource.Resource.
func (r *ModbusTcpSensor) Close(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logger.Info("Closing Modbus TCP Sensor Component")
	r.cancelFunc()
	if r.client != nil {
		err := r.client.Close()
		if err != nil {
			r.logger.Errorf("Failed to close modbus client: %#v", err)
		}
	}
	return nil
}

// DoCommand implements resource.Resource.
func (*ModbusTcpSensor) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{"ok": 1}, nil
}

// Reconfigure implements resource.Resource.
func (r *ModbusTcpSensor) Reconfigure(ctx context.Context, deps resource.Dependencies, conf resource.Config) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logger.Debug("Reconfiguring Modbus TCP Sensor Component")

	newConf, err := resource.NativeConfig[*ModbusTcpConfig](conf)
	if err != nil {
		return err
	}

	// In case the module has changed name
	r.Named = conf.ResourceName().AsNamed()

	return r.reconfigure(newConf, deps)
}

func (r *ModbusTcpSensor) reconfigure(newConf *ModbusTcpConfig, deps resource.Dependencies) error {
	r.logger.Infof("Reconfiguring Modbus TCP Sensor Component with %v", newConf)
	if r.client != nil {
		err := r.client.Close()
		if err != nil {
			r.logger.Errorf("Failed to close modbus client: %#v", err)
			// TODO: should we exit here?
		}
	}
	r.uri = newConf.Url
	r.timeout = time.Millisecond * time.Duration(newConf.Timeout)
	err := r.initializeModbusClient()
	if err != nil {
		return err
	}

	r.blocks = newConf.Blocks
	return nil
}

func (r *ModbusTcpSensor) initializeModbusClient() error {
	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     r.uri,
		Timeout: r.timeout,
	})
	if err != nil {
		r.logger.Errorf("Failed to create modbus client: %#v", err)
		return err
	}
	err = client.Open()
	if err != nil {
		r.logger.Errorf("Failed to open modbus client: %#v", err)
		return err
	}
	r.client = client
	return nil
}
