package clientbridge

import (
	"errors"
	"fmt"

	"viam-modbus/common"
)

type modbusBridgeConfig struct {
	Endpoints []common.ModbusConfig `json:"endpoints"`
}

func (cfg *modbusBridgeConfig) Validate(path string) ([]string, error) {
	if cfg.Endpoints == nil {
		return nil, errors.New("endpoints is required")
	}
	for i, endpoint := range cfg.Endpoints {
		if e := endpoint.Validate(); e != nil {
			return nil, fmt.Errorf("endpoint %v: %v", i, e)
		}
	}
	return nil, nil
}
