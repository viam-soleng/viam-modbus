package common

import (
	"errors"
	"net/url"
)

type ModbusConfig struct {
	Name          string `json:"name"`
	Endpoint      string `json:"endpoint"`
	ServerAddress uint16 `json:"server_id"`
	Speed         uint   `json:"speed"`
	DataBits      uint   `json:"data_bits"`
	Parity        string `json:"parity"`
	StopBits      uint   `json:"stop_bits"`
	// TLSClientCert string `json:"tls_client_cert"`
	// TLSRootCAs    string `json:"tls_root_cas"`
}

func (cfg *ModbusConfig) IsSerial() bool {
	u, _ := url.Parse(cfg.Endpoint)
	return u.Scheme == "ascii" || u.Scheme == "rtu"
}

func (cfg *ModbusConfig) IsRTU() bool {
	u, _ := url.Parse(cfg.Endpoint)
	return u.Scheme == "rtu"
}

func (cfg *ModbusConfig) IsNetwork() bool {
	u, _ := url.Parse(cfg.Endpoint)
	return u.Scheme == "tcp" || u.Scheme == "udp"
}

func (cfg *ModbusConfig) Validate() error {
	if cfg.Endpoint == "" {
		return errors.New("endpoint must be set")
	}

	u, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return err
	}
	switch u.Scheme {
	case "ascii":
		return cfg.ValidateSerialConfig(false)
	case "rtu":
		return cfg.ValidateSerialConfig(true)
	case "tcp":
		return cfg.ValidateTCPConfig()
	default:
		return errors.New("invalid endpoint scheme")
	}
}

func (cfg *ModbusConfig) ValidateSerialConfig(isRTU bool) error {
	if cfg.Speed == 0 {
		return errors.New("speed must be greater than 0")
	}

	if cfg.Parity != "N" && cfg.Parity != "E" && cfg.Parity != "O" {
		return errors.New("parity must be 'N', 'E', or 'O'")
	}

	if isRTU {
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

func (cfg *ModbusConfig) ValidateTCPConfig() error {
	return nil
}
