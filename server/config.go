package server

import (
	"errors"

	"viam-modbus/common"
)

type modbusBridgeConfig struct {
	Server      *common.ModbusConfig `json:"server"`
	PersistData bool                 `json:"persist_data"`
}

func (cfg *modbusBridgeConfig) Validate(path string) ([]string, error) {
	if cfg.Server == nil {
		return nil, errors.New("server is required")
	}
	if err := cfg.Server.Validate(); err != nil {
		return nil, err
	}
	return nil, nil
}
