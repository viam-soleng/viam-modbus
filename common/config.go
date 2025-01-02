package common

import (
	"errors"
	"net/url"
	"os"
)

var (
	ErrNoSuchDevice         = errors.New("no such device")
	ErrInvalidEndpoint      = errors.New("invalid endpoint")
	ErrInvalidLogLevel      = errors.New("invalid log level")
	ErrEndpointMustBeSet    = errors.New("endpoint must be set")
	ErrInvalidScheme        = errors.New("invalid scheme")
	ErrSpeedMustBeNonZero   = errors.New("speed must be non-zero")
	ErrInvalidParity        = errors.New("parity must be 'N', 'E', or 'O'")
	ErrInvalidRTUDataBits   = errors.New("data_bits must be 8 for RTU")
	ErrInvalidRTUStopBits   = errors.New("stop_bits must be 1 or 2")
	ErrInvalidASCIIDataBits = errors.New("data_bits must be 7 for ASCII")
	ErrInvalidASCIIStopBits = errors.New("stop_bits must be 1 for ASCII")
	ErrFailedToCreateClient = errors.New("failed to create client")
	ErrFailedToCreateServer = errors.New("failed to create server")
	ErrInvalidConfig        = errors.New("invalid config")
)

type ModbusConfig struct {
	Name          string `json:"name"`
	Endpoint      string `json:"endpoint"`
	ServerAddress uint16 `json:"server_id"`
	Speed         uint   `json:"speed"`
	DataBits      uint   `json:"data_bits"`
	Parity        string `json:"parity"`
	StopBits      uint   `json:"stop_bits"`
	LogLevel      string `json:"log_level"`
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
		return ErrEndpointMustBeSet
	}

	switch cfg.LogLevel {
	case "debug":
	case "info":
	case "warn":
	case "error":
	default:
		return ErrInvalidLogLevel
	}

	u, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return errors.Join(err, ErrInvalidEndpoint)
	}
	switch u.Scheme {
	case "ascii":
		return cfg.ValidateSerialConfig(false)
	case "rtu":
		return cfg.ValidateSerialConfig(true)
	case "tcp":
		return cfg.ValidateTCPConfig()
	default:
		return ErrInvalidScheme
	}
}

func (cfg *ModbusConfig) ValidateSerialConfig(isRTU bool) error {
	u, err := url.Parse(cfg.Endpoint)
	if err != nil {
		// This should never happen, as this is the second time we do the parse
		return errors.Join(err, ErrInvalidEndpoint)
	}

	if _, err := os.Stat(u.Path); os.IsNotExist(err) {
		return errors.Join(err, ErrNoSuchDevice)
	}

	if cfg.Speed == 0 {
		return ErrSpeedMustBeNonZero
	}

	if cfg.Parity != "N" && cfg.Parity != "E" && cfg.Parity != "O" {
		return ErrInvalidParity
	}

	if isRTU {
		if cfg.DataBits != 8 {
			return ErrInvalidRTUDataBits
		}

		if cfg.StopBits != 1 && cfg.StopBits != 2 {
			return ErrInvalidRTUStopBits
		}
	} else {
		if cfg.DataBits != 7 {
			return ErrInvalidASCIIDataBits
		}

		if cfg.StopBits != 1 {
			return ErrInvalidASCIIStopBits
		}
	}
	return nil
}

func (cfg *ModbusConfig) ValidateTCPConfig() error {
	return nil
}
