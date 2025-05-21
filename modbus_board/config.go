package modbus_board

import (
	"errors"
	"fmt"

	"viam-modbus/common"
)

type ModbusBoardCloudConfig struct {
	Modbus     *common.ModbusClientConfig   `json:"modbus"`
	GpioPins   []ModbusGpioPinCloudConfig   `json:"gpio_pins"`
	AnalogPins []ModbusAnalogPinCloudConfig `json:"analog_pins"`
}

type ModbusGpioPinCloudConfig struct {
	Offset  int    `json:"offset"`
	Name    string `json:"name"`
	PinType string `json:"pin_type"`
}

type ModbusAnalogPinCloudConfig struct {
	Offset   int    `json:"offset"`
	Name     string `json:"name"`
	PinType  string `json:"pin_type"`
	DataType string `json:"data_type"`
}


func (cfg *ModbusBoardCloudConfig) Validate(path string) ([]string, []string, error) {
	if cfg.Modbus == nil {
		return nil, nil, errors.New("modbus is required")
	}
	e := cfg.Modbus.Validate()
	if e != nil {
		return nil, nil, fmt.Errorf("modbus: %v", e)
	}

	if cfg.GpioPins != nil {
		for i, pin := range cfg.GpioPins {
			if pin.Name == "" {
				return nil, nil, fmt.Errorf("name is required in pin %d", i)
			}
			if pin.PinType == "" {
				return nil, nil, fmt.Errorf("type is required in pin %v", pin.Name)
			}
			if p := common.NewPinType(pin.PinType); p == common.UNKNOWN {
				return nil, nil, fmt.Errorf("type must be %v or %v in pin %v", common.OUTPUT_PIN, common.INPUT_PIN, pin.Name)
			}
			if pin.Offset < 0 {
				return nil, nil, fmt.Errorf("offset must be non-negative in pin %v", pin.Name)
			}
		}
	}

	if cfg.AnalogPins != nil {
		for i, pin := range cfg.AnalogPins {
			if pin.Name == "" {
				return nil, nil, fmt.Errorf("name is required in pin %d", i)
			}
			if pin.PinType == "" {
				return nil, nil, fmt.Errorf("type is required in pin %v", pin.Name)
			}
			if p := common.NewPinType(pin.PinType); p == common.UNKNOWN {
				return nil, nil, fmt.Errorf("type must be %v or %v in pin %v", common.OUTPUT_PIN, common.INPUT_PIN, pin.Name)
			}
			if pin.Offset < 0 {
				return nil, nil, fmt.Errorf("offset must be non-negative in pin %v", pin.Name)
			}
			if pin.DataType == "" {
				return nil, nil, fmt.Errorf("data_type is required in pin %v", pin.Name)
			}
		}
	}
	return nil, nil, nil
}
