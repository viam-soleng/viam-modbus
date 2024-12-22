package serverbridge

import (
	"errors"
	"fmt"

	"viam-modbus/common"
)

type modbusBridgeConfig struct {
	Endpoints []namedModbusConfig `json:"endpoints"`
}

type namedModbusConfig struct {
	common.ModbusConfig
	Name string `json:"name"`
}

func (cfg *modbusBridgeConfig) Validate(path string) ([]string, error) {
	if cfg.Endpoints == nil {
		return nil, errors.New("endpoints is required")
	}
	existingNames := make([]string, len(cfg.Endpoints))
	for i, endpoint := range cfg.Endpoints {
		if e := endpoint.Validate(); e != nil {
			return nil, fmt.Errorf("endpoint %v: %v", i, e)
		}
		for _, name := range existingNames {
			if name == endpoint.Name {
				return nil, fmt.Errorf("duplicate endpoint name: %v", endpoint.Name)
			}
		}
		existingNames[i] = endpoint.Name
	}
	return nil, nil
}
