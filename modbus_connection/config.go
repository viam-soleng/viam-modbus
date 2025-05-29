package modbus_connection

import (
	"errors"
	"fmt"

	"viam-modbus/common"
)

type ModbusConnectionConfig struct {
	Modbus *common.ModbusClientConfig `json:"modbus"`
}


func (cfg *ModbusConnectionConfig) Validate(path string) ([]string, []string, error) {
	if cfg.Modbus == nil {
		return nil, nil, errors.New("modbus is required")
	}
	e := cfg.Modbus.Validate()
	if e != nil {
		return nil, nil, fmt.Errorf("modbus: %v", e)
	}

	return nil, nil, nil
}
