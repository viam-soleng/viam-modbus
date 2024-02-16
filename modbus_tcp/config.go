package modbus_tcp

import (
	"errors"
	"fmt"
)

type ModbusTcpConfig struct {
	// use "tcp://host:port" format for TCP
	// use "udp://device:port" format for UDP
	Url     string         `json:"url"`
	Timeout int            `json:"timeout_ms"`
	Blocks  []ModbusBlocks `json:"blocks"`
}

type ModbusBlocks struct {
	Offset int    `json:"offset"`
	Length int    `json:"length"`
	Type   string `json:"type"`
	Name   string `json:"name"`
}

func (cfg *ModbusTcpConfig) Validate(path string) ([]string, error) {
	if cfg.Url == "" {
		return nil, errors.New("url is required")
	}
	if cfg.Timeout < 0 {
		return nil, errors.New("timeout must be non-negative")
	}
	if len(cfg.Blocks) == 0 {
		return nil, errors.New("blocks must be non-empty")
	}
	for i, block := range cfg.Blocks {
		// TODO: handle special case where type doesn't require a length
		if block.Offset < 0 {
			return nil, fmt.Errorf("offset must be non-negative in block %v", i)
		}
		if block.Length <= 0 {
			return nil, fmt.Errorf("length must be non-zero and non-negative in block %v", i)
		}
	}
	return nil, nil
}
