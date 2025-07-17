package viammodbus

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/simonvetter/modbus"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
)

var ModbusSensorModel = NamespaceFamily.WithModel("sensor")

// Registers the sensor model
func init() {
	resource.RegisterComponent(
		sensor.API,
		ModbusSensorModel,
		resource.Registration[sensor.Sensor, *ModbusSensorConfig]{
			Constructor: NewModbusSensor,
		},
	)
}

type ModbusSensorConfig struct {
	ModbusConnection string         `json:"modbus_connection_name"`
	Blocks           []ModbusBlocks `json:"blocks"`
	UnitID           int            `json:"unit_id,omitempty"` // Optional unit ID for Modbus commands
}

type ModbusBlocks struct {
	Offset int    `json:"offset"`
	Length int    `json:"length"`
	Type   string `json:"type"`
	Name   string `json:"name"`
}

func (cfg *ModbusSensorConfig) Validate(path string) ([]string, []string, error) {
	if cfg.ModbusConnection == "" {
		return nil, nil, resource.NewConfigValidationFieldRequiredError(path, "modbus_connection_name")
	}

	if cfg.Blocks == nil {
		return nil, nil, errors.New("blocks is required")
	}

	for i, block := range cfg.Blocks {
		if block.Name == "" {
			return nil, nil, fmt.Errorf("name is required in block %v", i)
		}
		if block.Type == "" {
			return nil, nil, fmt.Errorf("type is required in block %v", i)
		}
		if block.Offset < 0 {
			return nil, nil, fmt.Errorf("offset must be non-negative in block %v", i)
		}
		if shouldCheckLength(block.Type) && block.Length <= 0 {
			return nil, nil, fmt.Errorf("length must be non-zero and non-negative in block %v", i)
		}
	}

	if cfg.UnitID < 0 || cfg.UnitID > 247 {
		return nil, nil, fmt.Errorf("unit_id must be between 0 and 247 or removed, got %d", cfg.UnitID)
	}
	return []string{string(cfg.ModbusConnection)}, nil, nil
}

func shouldCheckLength(t string) bool {
	switch t {
	case "coils", "discrete_inputs", "holding_registers", "input_registers", "bytes", "rawBytes":
		return true
	default:
		return false
	}
}

// Creates a new modbus sensor instance
func NewModbusSensor(ctx context.Context, deps resource.Dependencies, conf resource.Config, logger logging.Logger) (sensor.Sensor, error) {
	newConf, err := resource.NativeConfig[*ModbusSensorConfig](conf)
	if err != nil {
		return nil, err
	}

	c, cancelFunc := context.WithCancel(context.Background())
	s := ModbusSensor{
		Named:      conf.ResourceName().AsNamed(),
		logger:     logger,
		cancelFunc: cancelFunc,
		ctx:        c,
		blocks:     newConf.Blocks,
	}

	client, err := GlobalClientRegistry.Get(newConf.ModbusConnection)
	if err != nil {
		return nil, err
	}
	s.mc = client
	return &s, nil
}

type ModbusSensor struct {
	resource.AlwaysRebuild
	resource.Named
	mu         sync.RWMutex
	logger     logging.Logger
	cancelFunc context.CancelFunc
	ctx        context.Context
	blocks     []ModbusBlocks
	unitID     *uint8 // Optional unit ID for Modbus commands
	mc         *modbusClient
}

// Returns modbus register values
func (s *ModbusSensor) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.mc == nil {
		return nil, errors.New("modbus client not initialized")
	}
	results := map[string]interface{}{}
	for _, block := range s.blocks {
		switch block.Type {
		case "coils":
			b, err := s.mc.ReadCoils(uint16(block.Offset), uint16(block.Length), s.unitID)
			if err != nil {
				return nil, err
			}
			writeBoolArrayToOutput(b, block, results)
			s.mc.logger.Warnf("Read %d coils starting at offset %d", len(b), block.Offset)
		case "discrete_inputs":
			b, err := s.mc.ReadDiscreteInputs(uint16(block.Offset), uint16(block.Length), s.unitID)
			if err != nil {
				return nil, err
			}
			writeBoolArrayToOutput(b, block, results)
		case "holding_registers":
			b, err := s.mc.ReadHoldingRegisters(uint16(block.Offset), uint16(block.Length), s.unitID)
			if err != nil {
				return nil, err
			}
			writeUInt16ArrayToOutput(b, block, results)
		case "input_registers":
			b, err := s.mc.ReadInputRegisters(uint16(block.Offset), uint16(block.Length), s.unitID)
			if err != nil {
				return nil, err
			}
			writeUInt16ArrayToOutput(b, block, results)
		case "bytes":
			b, e := s.mc.ReadBytes(uint16(block.Offset), uint16(block.Length), modbus.HOLDING_REGISTER, s.unitID)
			if e != nil {
				return nil, e
			}
			writeByteArrayToOutput(b, block, results)
		case "rawBytes":
			b, e := s.mc.ReadRawBytes(uint16(block.Offset), uint16(block.Length), modbus.HOLDING_REGISTER, s.unitID)
			if e != nil {
				return nil, e
			}
			writeByteArrayToOutput(b, block, results)
		case "uint8":
			b, e := s.mc.ReadUInt8(uint16(block.Offset), modbus.HOLDING_REGISTER, s.unitID)
			if e != nil {
				return nil, e
			}
			results[block.Name] = int32(b)
		case "int16":
			b, e := s.mc.ReadInt16(uint16(block.Offset), modbus.HOLDING_REGISTER, s.unitID)
			if e != nil {
				return nil, e
			}
			results[block.Name] = int32(b)
		case "uint16":
			b, e := s.mc.ReadUInt16(uint16(block.Offset), modbus.HOLDING_REGISTER, s.unitID)
			if e != nil {
				return nil, e
			}
			results[block.Name] = int32(b)
		case "int32":
			b, e := s.mc.ReadInt32(uint16(block.Offset), modbus.HOLDING_REGISTER, s.unitID)
			if e != nil {
				return nil, e
			}
			results[block.Name] = b
		case "uint32":
			b, e := s.mc.ReadUInt32(uint16(block.Offset), modbus.HOLDING_REGISTER, s.unitID)
			if e != nil {
				return nil, e
			}
			results[block.Name] = b
		case "float32":
			b, e := s.mc.ReadFloat32(uint16(block.Offset), modbus.HOLDING_REGISTER, s.unitID)
			if e != nil {
				return nil, e
			}
			results[block.Name] = b
		case "float64":
			b, e := s.mc.ReadFloat64(uint16(block.Offset), modbus.HOLDING_REGISTER, s.unitID)
			if e != nil {
				return nil, e
			}
			results[block.Name] = b
		default:
			results[block.Name] = "unsupported type"
		}
	}
	s.logger.Infof("Readings for sensor %s: %v", s.Named.Name, results)
	return results, nil
}

// DoCommand currently not implemented
func (*ModbusSensor) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{}, fmt.Errorf("DoCommand not implemented for ModbusSensor")
}

// Closes the modbus sensor instance
func (s *ModbusSensor) Close(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logger.Info("Closing Modbus Sensor Component")
	s.cancelFunc()

	return nil
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
