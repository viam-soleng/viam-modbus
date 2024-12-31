package common

import (
	"errors"
	"net/url"
	"time"

	client "github.com/rinzlerlabs/gomodbus/client"
	network_client "github.com/rinzlerlabs/gomodbus/client/network"
	ascii_client "github.com/rinzlerlabs/gomodbus/client/serial/ascii"
	rtu_client "github.com/rinzlerlabs/gomodbus/client/serial/rtu"
	server "github.com/rinzlerlabs/gomodbus/server"
	network_server "github.com/rinzlerlabs/gomodbus/server/network"
	ascii_server "github.com/rinzlerlabs/gomodbus/server/serial/ascii"
	rtu_server "github.com/rinzlerlabs/gomodbus/server/serial/rtu"
	network_settings "github.com/rinzlerlabs/gomodbus/settings/network"
	serial_settings "github.com/rinzlerlabs/gomodbus/settings/serial"
	"go.viam.com/rdk/logging"
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

func NewModbusClientFromConfig(logger logging.Logger, conf *ModbusConfig) (client.ModbusClient, error) {
	endpoint, err := url.Parse(conf.Endpoint)
	if err != nil {
		return nil, err
	}
	if conf.IsSerial() {
		settings := &serial_settings.ClientSettings{
			SerialSettings: serial_settings.SerialSettings{
				Device:   endpoint.Path,
				Baud:     int(conf.Speed),
				DataBits: int(conf.DataBits),
				Parity:   conf.Parity,
				StopBits: int(conf.StopBits),
			},
			ResponseTimeout: 1 * time.Second,
		}
		if conf.IsRTU() {
			client, err := rtu_client.NewModbusClientFromSettings(logger.AsZap().Desugar(), settings)
			if err != nil {
				return nil, errors.Join(err, errors.New("failed to create rtu client"))
			}
			return client, nil
		} else {
			client, err := ascii_client.NewModbusClientFromSettings(logger.AsZap().Desugar(), settings)
			if err != nil {
				return nil, errors.Join(err, errors.New("failed to create ascii client"))
			}
			return client, nil
		}
	} else if conf.IsNetwork() {
		settings := &network_settings.ClientSettings{
			NetworkSettings: network_settings.NetworkSettings{
				Endpoint:  endpoint,
				KeepAlive: 30 * time.Second,
			},
			ResponseTimeout: 1 * time.Second,
			DialTimeout:     1 * time.Second,
		}
		client, err := network_client.NewModbusClientFromSettings(logger.AsZap().Desugar(), settings)
		if err != nil {
			return nil, errors.Join(err, errors.New("failed to create network client"))
		}
		return client, nil
	} else {
		return nil, errors.New("invalid config")
	}
}

func NewModbusServerFromConfigWithHandler(logger logging.Logger, conf *ModbusConfig, handler server.RequestHandler) (server.ModbusServer, error) {
	u, err := url.Parse(conf.Endpoint)
	if err != nil {
		return nil, err
	}
	if conf.IsSerial() {
		settings := &serial_settings.ServerSettings{
			SerialSettings: serial_settings.SerialSettings{
				Device:   u.Path,
				Baud:     int(conf.Speed),
				DataBits: int(conf.DataBits),
				Parity:   conf.Parity,
				StopBits: int(conf.StopBits),
			},
			Address: conf.ServerAddress,
		}
		if conf.IsRTU() {
			var server server.ModbusServer
			var err error
			if handler != nil {
				server, err = rtu_server.NewModbusServerWithHandler(logger.Desugar(), settings, handler)
			} else {
				server, err = rtu_server.NewModbusServerFromSettings(logger.Desugar(), settings)
			}
			if err != nil {
				return nil, errors.Join(err, errors.New("failed to create rtu server"))
			}
			return server, nil
		} else {
			var server server.ModbusServer
			var err error
			if handler != nil {
				server, err = ascii_server.NewModbusServerWithHandler(logger.Desugar(), settings, handler)
			} else {
				server, err = ascii_server.NewModbusServerFromSettings(logger.Desugar(), settings)
			}
			if err != nil {
				return nil, errors.Join(err, errors.New("failed to create ascii server"))
			}
			return server, nil
		}
	} else if conf.IsNetwork() {
		settings := &network_settings.ServerSettings{
			NetworkSettings: network_settings.NetworkSettings{
				Endpoint:  u,
				KeepAlive: 30 * time.Second,
			},
		}
		var server server.ModbusServer
		var err error
		if handler != nil {
			server, err = network_server.NewModbusServerWithHandler(logger.Desugar(), settings, handler)
		} else {
			server, err = network_server.NewModbusServerFromSettings(logger.Desugar(), settings)
		}
		if err != nil {
			return nil, errors.Join(err, errors.New("failed to create network server"))
		}
		return server, nil
	} else {
		return nil, errors.New("invalid config")
	}
}

func NewModbusServerFromConfig(conf *ModbusConfig, logger logging.Logger) (server.ModbusServer, error) {
	return NewModbusServerFromConfigWithHandler(logger, conf, nil)
}
