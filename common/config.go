package common

import (
	"errors"
)

type ModbusConfig struct {
	Name         string              `json:"name"`
	Endpoint     string              `json:"endpoint"`
	SerialConfig *ModbusSerialConfig `json:"serial_config"`
	TCPConfig    *ModbusTCPConfig    `json:"tcp_config"`
	// Timeout      int                 `json:"timeout_ms"`
}

type ModbusSerialConfig struct {
	ServerAddress uint16 `json:"server_id"`
	Speed         uint   `json:"speed"`
	DataBits      uint   `json:"data_bits"`
	Parity        string `json:"parity"`
	StopBits      uint   `json:"stop_bits"`
	RTU           bool   `json:"rtu"`
}

type ModbusTCPConfig struct {
	// TLSClientCert string `json:"tls_client_cert"`
	// TLSRootCAs    string `json:"tls_root_cas"`
}

func (cfg *ModbusConfig) Validate() error {
	// if cfg.Timeout <= 0 {
	// 	return errors.New("timeout_ms must be greater than 0")
	// }

	if cfg.Endpoint == "" {
		return errors.New("endpoint must be set")
	}

	if cfg.SerialConfig != nil && cfg.TCPConfig != nil {
		return errors.New("only one of serial_config or tcp_config can be set")
	}

	if cfg.SerialConfig != nil {
		return cfg.SerialConfig.Validate()
	}

	if cfg.TCPConfig != nil {
		return cfg.TCPConfig.Validate()
	}

	return errors.New("either serial_config or tcp_config must be set")
}

func (cfg *ModbusSerialConfig) Validate() error {
	if cfg.Speed == 0 {
		return errors.New("speed must be greater than 0")
	}

	if cfg.Parity != "N" && cfg.Parity != "E" && cfg.Parity != "O" {
		return errors.New("parity must be 'N', 'E', or 'O'")
	}

	if cfg.RTU {
		if cfg.DataBits != 8 {
			return errors.New("data_bits must be 8 for RTU")
		}

		if cfg.StopBits != 1 && cfg.StopBits != 2 {
			return errors.New("stop_bits must be 1 or 2 for RTU")
		}
	} else {
		if cfg.DataBits != 7 {
			return errors.New("data_bits must be 7 for ASCII")
		}

		if cfg.StopBits != 1 {
			return errors.New("stop_bits must be 1 for ASCII")
		}
	}
	return nil
}

func (cfg *ModbusTCPConfig) Validate() error {
	return nil
}
