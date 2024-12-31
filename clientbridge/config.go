package clientbridge

import (
	"errors"
	"fmt"

	"viam-modbus/common"
)

type modbusBridgeConfig struct {
	Endpoints    []*common.ModbusConfig `json:"endpoints"`
	UpdateTimeMs uint                   `json:"update_time_ms"`
	Blocks       []modbusBlock          `json:"blocks"`
}

type modbusBlock struct {
	Src         string `json:"src"`
	SrcOffset   int    `json:"src_offset"`
	SrcRegister string `json:"src_register"`
	Dst         string `json:"dst"`
	DstOffset   int    `json:"dst_offset"`
	DstRegister string `json:"dst_register"`
	Length      int    `json:"length"`
}

func (cfg *modbusBridgeConfig) Validate(path string) ([]string, error) {
	if cfg.Endpoints == nil {
		return nil, errors.New("endpoints is required")
	}
	if cfg.UpdateTimeMs == 0 {
		return nil, errors.New("update_time_ms must be greater than 0")
	}
	endpointNames := make([]string, len(cfg.Endpoints))
	for i, endpoint := range cfg.Endpoints {
		if e := endpoint.Validate(); e != nil {
			return nil, fmt.Errorf("endpoint %v: %v", i, e)
		}
		for _, name := range endpointNames {
			if name == endpoint.Name {
				return nil, fmt.Errorf("duplicate endpoint name: %v", endpoint.Name)
			}
		}
		endpointNames[i] = endpoint.Name
	}
	for i, block := range cfg.Blocks {
		if block.Src == "" {
			return nil, fmt.Errorf("src is required in block %v", i)
		}
		if !IsIn(block.Src, endpointNames) {
			return nil, fmt.Errorf("src %v not found in endpoints", block.Src)
		}
		if block.Dst == "" {
			return nil, fmt.Errorf("dst is required in block %v", i)
		}
		if !IsIn(block.Dst, endpointNames) {
			return nil, fmt.Errorf("dst %v not found in endpoints", block.Dst)
		}
		if block.SrcRegister == "" {
			return nil, fmt.Errorf("src_register is required in block %v", i)
		}
		if block.DstRegister == "" {
			return nil, fmt.Errorf("dst_register is required in block %v", i)
		}
		if block.SrcRegister == "coils" && block.DstRegister != "coils" {
			return nil, fmt.Errorf("src_register is coils, dst_register must be coils in block %v", i)
		}
		if block.SrcRegister == "discrete_inputs" && block.DstRegister != "coils" {
			return nil, fmt.Errorf("src_register is discrete_inputs, dst_register must be coils in block %v", i)
		}
		if block.SrcRegister == "holding_registers" && block.DstRegister != "holding_registers" {
			return nil, fmt.Errorf("src_register is holding_registers, dst_register must be holding_registers in block %v", i)
		}
		if block.SrcRegister == "input_registers" && block.DstRegister != "holding_registers" {
			return nil, fmt.Errorf("src_register is input_registers, dst_register must be holding_registers in block %v", i)
		}
		if block.SrcOffset < 0 {
			return nil, fmt.Errorf("src_offset must be non-negative in block %v", i)
		}
		if block.DstOffset < 0 {
			return nil, fmt.Errorf("dst_offset must be non-negative in block %v", i)
		}
		if block.Length <= 0 {
			return nil, fmt.Errorf("length must be non-zero and non-negative in block %v", i)
		}
	}
	return nil, nil
}

func IsIn(val string, arr []string) bool {
	for _, el := range arr {
		if el == val {
			return true
		}
	}
	return false
}
