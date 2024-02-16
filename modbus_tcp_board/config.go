package modbus_tcp_board

import (
	"errors"
	"fmt"
)

type ModbusTcpBoardConfig struct {
	// use "tcp://host:port" format for TCP
	// use "udp://device:port" format for UDP
	Url        string                  `json:"url"`
	Timeout    int                     `json:"timeout_ms"`
	GpioPins   []ModbusGpioPinConfig   `json:"gpio_pins"`
	AnalogPins []ModbusAnalogPinConfig `json:"analog_pins"`
	Endianness string                  `json:"endianness"`
	WordOrder  string                  `json:"word_order"`
}

type ModbusGpioPinConfig struct {
	Offset  int    `json:"offset"`
	Name    string `json:"name"`
	PinType string `json:"pin_type"`
}

type ModbusAnalogPinConfig struct {
	Offset   int    `json:"offset"`
	Name     string `json:"name"`
	PinType  string `json:"pin_type"`
	DataType string `json:"data_type"`
}

func (cfg *ModbusTcpBoardConfig) Validate(path string) ([]string, error) {
	if cfg.Url == "" {
		return nil, errors.New("url is required")
	}
	if cfg.Timeout < 0 {
		return nil, errors.New("timeout must be non-negative")
	}
	if cfg.GpioPins != nil {
		for i, pin := range cfg.GpioPins {
			if pin.Name == "" {
				return nil, fmt.Errorf("name is required in pin %d", i)
			}
			if pin.PinType == "" {
				return nil, fmt.Errorf("type is required in pin %v", pin.Name)
			}
			if pin.PinType != OUTPUT_PIN && pin.PinType != INPUT_PIN {
				return nil, fmt.Errorf("type must be %v or %v in pin %v", OUTPUT_PIN, INPUT_PIN, pin.Name)
			}
			if pin.Offset < 0 {
				return nil, fmt.Errorf("offset must be non-negative in pin %v", pin.Name)
			}
		}
	}
	return nil, nil
}
