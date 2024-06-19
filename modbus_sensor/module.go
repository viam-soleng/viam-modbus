package modbus_sensor

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/simonvetter/modbus"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"

	"viam-modbus/common"
	"viam-modbus/utils"
)

// TODO: change model from "modbus-tcp" to "modbus"
var Model = resource.NewModel("viam-soleng", "sensor", "modbus-tcp")

func init() {
	resource.RegisterComponent(
		sensor.API,
		Model,
		resource.Registration[sensor.Sensor, *ModbusSensorConfig]{
			Constructor: NewModbusSensor,
		},
	)
}

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
	mu         sync.RWMutex
	logger     logging.Logger
	cancelFunc context.CancelFunc
	ctx        context.Context
	client     *common.ModbusClient
	blocks     []ModbusBlocks
}

// Readings implements sensor.Sensor.
func (r *ModbusSensor) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.client == nil {
		return nil, errors.New("modbus client not initialized")
	}
	results := map[string]interface{}{}
	for _, block := range r.blocks {
		switch block.Type {
		case "coils":
			b, err := r.client.ReadCoils(uint16(block.Offset), uint16(block.Length))
			if err != nil {
				return nil, err
			}
			writeBoolArrayToOutput(b, block, results)
		case "discrete_inputs":
			b, err := r.client.ReadDiscreteInputs(uint16(block.Offset), uint16(block.Length))
			if err != nil {
				return nil, err
			}
			writeBoolArrayToOutput(b, block, results)
		case "holding_registers":
			b, err := r.client.ReadHoldingRegisters(uint16(block.Offset), uint16(block.Length))
			if err != nil {
				return nil, err
			}
			writeUInt16ArrayToOutput(b, block, results)
		case "input_registers":
			b, err := r.client.ReadInputRegisters(uint16(block.Offset), uint16(block.Length))
			if err != nil {
				return nil, err
			}
			writeUInt16ArrayToOutput(b, block, results)
		case "bytes":
			b, e := r.client.ReadBytes(uint16(block.Offset), uint16(block.Length), modbus.HOLDING_REGISTER)
			if e != nil {
				return nil, e
			}
			writeByteArrayToOutput(b, block, results)
		case "rawBytes":
			b, e := r.client.ReadRawBytes(uint16(block.Offset), uint16(block.Length), modbus.HOLDING_REGISTER)
			if e != nil {
				return nil, e
			}
			writeByteArrayToOutput(b, block, results)
		case "uint8":
			b, e := r.client.ReadUInt8(uint16(block.Offset), modbus.HOLDING_REGISTER)
			if e != nil {
				return nil, e
			}
			results[block.Name] = int32(b)
		case "int16":
			b, e := r.client.ReadInt16(uint16(block.Offset), modbus.HOLDING_REGISTER)
			if e != nil {
				return nil, e
			}
			results[block.Name] = int32(b)
		case "uint16":
			b, e := r.client.ReadUInt16(uint16(block.Offset), modbus.HOLDING_REGISTER)
			if e != nil {
				return nil, e
			}
			results[block.Name] = int32(b)
		case "int32":
			b, e := r.client.ReadInt32(uint16(block.Offset), modbus.HOLDING_REGISTER)
			if e != nil {
				return nil, e
			}
			results[block.Name] = b
		case "uint32":
			b, e := r.client.ReadUInt32(uint16(block.Offset), modbus.HOLDING_REGISTER)
			if e != nil {
				return nil, e
			}
			results[block.Name] = b
		case "float32":
			b, e := r.client.ReadFloat32(uint16(block.Offset), modbus.HOLDING_REGISTER)
			if e != nil {
				return nil, e
			}
			results[block.Name] = b
		case "float64":
			b, e := r.client.ReadFloat64(uint16(block.Offset), modbus.HOLDING_REGISTER)
			if e != nil {
				return nil, e
			}
			results[block.Name] = b
		default:
			results[block.Name] = "unsupported type"
		}
	}
	return results, nil
}

func writeBoolArrayToOutput(b []bool, block ModbusBlocks, results map[string]interface{}) {
	for i, v := range b {
		field_name := block.Name + "_" + fmt.Sprint(i)
		results[field_name] = v
	}
}

func writeUInt16ArrayToOutput(b []uint16, block ModbusBlocks, results map[string]interface{}) {
	for i, v := range b {
		field_name := block.Name + "_" + fmt.Sprint(i)
		results[field_name] = strconv.Itoa(int(v))
	}
}

func writeByteArrayToOutput(b []byte, block ModbusBlocks, results map[string]interface{}) {
	results[block.Name] = hex.EncodeToString(b)
}

// Close implements resource.Resource.
func (r *ModbusSensor) Close(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logger.Info("Closing Modbus Sensor Component")
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
func (*ModbusSensor) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{"ok": 1}, nil
}

// Reconfigure implements resource.Resource.
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

func (r *ModbusSensor) reconfigure(newConf *ModbusSensorConfig, deps resource.Dependencies) error {
	r.logger.Infof("Reconfiguring Modbus Sensor Component with %v", newConf)
	if r.client != nil {
		err := r.client.Close()
		if err != nil {
			r.logger.Errorf("Failed to close modbus client: %#v", err)
			// TODO: should we exit here?
		}
	}

	endianness, err := common.GetEndianness(newConf.Modbus.Endianness)
	if err != nil {
		return err
	}

	wordOrder, err := common.GetWordOrder(newConf.Modbus.WordOrder)
	if err != nil {
		return err
	}

	timeout := time.Millisecond * time.Duration(newConf.Modbus.Timeout)
	client, err := common.NewModbusClient(r.logger, newConf.Modbus.Url, timeout, endianness, wordOrder, newConf.Modbus.Speed)
	if err != nil {
		return err
	}
	r.client = client

	r.blocks = newConf.Blocks
	return nil
}
