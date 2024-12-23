package serverbridge

import (
	"errors"
	"fmt"

	"viam-modbus/common"
)

type modbusBridgeConfig struct {
	Servers     []common.ModbusConfig `json:"servers"`
	PersistData bool                  `json:"persist_data"`
}

func (cfg *modbusBridgeConfig) Validate(path string) ([]string, error) {
	if cfg.Servers == nil {
		return nil, errors.New("servers is required")
	}
	existingNames := make([]string, len(cfg.Servers))
	for i, serverCfg := range cfg.Servers {
		if e := serverCfg.Validate(); e != nil {
			return nil, fmt.Errorf("server %d: %v", i, e)
		}
		for _, name := range existingNames {
			if name == serverCfg.Name {
				return nil, fmt.Errorf("duplicate endpoint name: %v", serverCfg.Name)
			}
		}
		existingNames[i] = serverCfg.Name
	}
	return nil, nil
}
