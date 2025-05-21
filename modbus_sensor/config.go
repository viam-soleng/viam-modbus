package modbus_sensor

import (
	"errors"
	"fmt"

	"viam-modbus/common"
)

type ModbusSensorConfig struct {
	Modbus *common.ModbusClientConfig `json:"modbus"`
	Blocks []ModbusBlocks             `json:"blocks"`
}

type ModbusBlocks struct {
	Offset int    `json:"offset"`
	Length int    `json:"length"`
	Type   string `json:"type"`
	Name   string `json:"name"`
}


func (cfg *ModbusSensorConfig) Validate(path string) ([]string, []string, error) {
	if cfg.Modbus == nil {
		return nil, nil, errors.New("modbus is required")
	}
	//TODO: Add TCP and RTU configuration validation
	e := cfg.Modbus.Validate()
	if e != nil {
		return nil, nil, fmt.Errorf("modbus: %v", e)
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
	return nil, nil, nil
}


func shouldCheckLength(t string) bool {
	switch t {
	case "coils", "discrete_inputs", "holding_registers", "input_registers", "bytes", "rawBytes":
		return true
	default:
		return false
	}
}
