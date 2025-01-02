// Package common provides common functionality across all modbus devices in this module
package common

import (
	"errors"
	"math"
	"net/url"
	"sync"
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
	"go.uber.org/zap"
	"go.viam.com/rdk/logging"
)

type Endianness int
type WordOrder int
type RegisterType int

const (
	BigEndian       Endianness   = 0
	LittleEndian    Endianness   = 1
	HighWordFirst   WordOrder    = 0
	LowWordFirst    WordOrder    = 1
	HoldingRegister RegisterType = 0
	InputRegister   RegisterType = 1
)

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
				return nil, errors.Join(err, ErrFailedToCreateClient)
			}
			return client, nil
		} else {
			client, err := ascii_client.NewModbusClientFromSettings(logger.AsZap().Desugar(), settings)
			if err != nil {
				return nil, errors.Join(err, ErrFailedToCreateClient)
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
			return nil, errors.Join(err, ErrFailedToCreateClient)
		}
		return client, nil
	} else {
		return nil, ErrInvalidConfig
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
				return nil, errors.Join(err, ErrFailedToCreateServer)
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
				return nil, errors.Join(err, ErrFailedToCreateServer)
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
			return nil, errors.Join(err, ErrFailedToCreateServer)
		}
		return server, nil
	} else {
		return nil, ErrInvalidConfig
	}
}

func NewModbusServerFromConfig(conf *ModbusConfig, logger logging.Logger) (server.ModbusServer, error) {
	return NewModbusServerFromConfigWithHandler(logger, conf, nil)
}

type ViamModbusClientWithRetry struct {
	mu            sync.RWMutex
	logger        logging.Logger
	conf          *ModbusConfig
	modbusClient  client.ModbusClient
	serverAddress uint16
}

func NewModbusClient(logger logging.Logger, conf *ModbusConfig) (*ViamModbusClientWithRetry, error) {
	client := &ViamModbusClientWithRetry{
		logger: logger,
	}
	err := client.initializeModbusClient()
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (r *ViamModbusClientWithRetry) Close() error {
	if r.modbusClient != nil {
		r.modbusClient.Close()
	}
	return nil
}

func (r *ViamModbusClientWithRetry) reinitializeModbusClient() error {
	r.logger.Warnf("Re-initializing modbus client")
	return r.initializeModbusClient()
}

func (r *ViamModbusClientWithRetry) initializeModbusClient() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.modbusClient != nil {
		err := r.modbusClient.Close()
		if err != nil {
			r.logger.Error("failed to close modbus client", zap.Error(err))
		}
	}

	client, err := NewModbusClientFromConfig(r.logger, r.conf)
	if err != nil {
		return err
	}
	r.modbusClient = client
	return nil
}

func (r *ViamModbusClientWithRetry) ReadCoils(offset, length uint16) ([]bool, error) {
	availableRetries := 3

	for availableRetries > 0 {
		b, err := r.modbusClient.ReadCoils(r.serverAddress, offset, length)
		if err != nil {
			availableRetries--
			err := r.reinitializeModbusClient()
			if err != nil {
				return nil, err
			}
		} else {
			return b, nil
		}
	}
	return nil, ErrRetriesExhausted
}

func (r *ViamModbusClientWithRetry) ReadCoil(offset uint16) (bool, error) {
	availableRetries := 3

	for availableRetries > 0 {
		b, err := r.modbusClient.ReadCoils(r.serverAddress, offset, 1)
		if err != nil {
			availableRetries--
			err := r.reinitializeModbusClient()
			if err != nil {
				return false, err
			}
		} else {
			return b[0], nil
		}
	}
	return false, ErrRetriesExhausted
}

func (r *ViamModbusClientWithRetry) ReadDiscreteInputs(offset, length uint16) ([]bool, error) {
	availableRetries := 3

	for availableRetries > 0 {
		b, err := r.modbusClient.ReadDiscreteInputs(r.serverAddress, offset, length)
		if err != nil {
			availableRetries--
			err := r.reinitializeModbusClient()
			if err != nil {
				return nil, err
			}
		} else {
			return b, nil
		}
	}
	return nil, ErrRetriesExhausted
}

func (r *ViamModbusClientWithRetry) ReadDiscreteInput(offset uint16) (bool, error) {
	availableRetries := 3

	for availableRetries > 0 {
		b, err := r.modbusClient.ReadDiscreteInputs(r.serverAddress, offset, 1)
		if err != nil {
			availableRetries--
			err := r.reinitializeModbusClient()
			if err != nil {
				return false, err
			}
		} else {
			return b[0], nil
		}
	}
	return false, ErrRetriesExhausted
}

func (r *ViamModbusClientWithRetry) WriteCoil(offset uint16, value bool) error {
	availableRetries := 3

	for availableRetries > 0 {
		err := r.modbusClient.WriteSingleCoil(r.serverAddress, offset, value)
		if err != nil {
			availableRetries--
			err := r.reinitializeModbusClient()
			if err != nil {
				return err
			}
		} else {
			return nil
		}
	}
	return ErrRetriesExhausted
}

func (r *ViamModbusClientWithRetry) ReadHoldingRegisters(offset, length uint16) ([]uint16, error) {
	availableRetries := 3

	for availableRetries > 0 {
		b, err := r.modbusClient.ReadHoldingRegisters(r.serverAddress, offset, length)
		if err != nil {
			availableRetries--
			err := r.reinitializeModbusClient()
			if err != nil {
				return nil, err
			}
		} else {
			return b, nil
		}
	}
	return nil, ErrRetriesExhausted
}

func (r *ViamModbusClientWithRetry) ReadInputRegisters(offset, length uint16) ([]uint16, error) {
	availableRetries := 3

	for availableRetries > 0 {
		b, err := r.modbusClient.ReadInputRegisters(r.serverAddress, offset, length)
		if err != nil {
			availableRetries--
			err := r.reinitializeModbusClient()
			if err != nil {
				return nil, err
			}
		} else {
			return b, nil
		}
	}
	return nil, ErrRetriesExhausted
}

// TODO: merge this with executeOperationWithRetry
func (r *ViamModbusClientWithRetry) rawRead(offset, len uint16, regType RegisterType) ([]uint16, error) {
	availableRetries := 3

	for availableRetries > 0 {
		var b []uint16
		var err error
		if regType == HoldingRegister {
			b, err = r.modbusClient.ReadHoldingRegisters(r.serverAddress, offset, len)
		} else if regType == InputRegister {
			b, err = r.modbusClient.ReadHoldingRegisters(r.serverAddress, offset, len)
		} else {
			return nil, ErrInvalidRegisterType
		}
		if err != nil {
			availableRetries--
			err := r.reinitializeModbusClient()
			if err != nil {
				return nil, err
			}
		} else {
			return b, nil
		}
	}
	return nil, ErrRetriesExhausted
}

func (r *ViamModbusClientWithRetry) ReadInt8(offset uint16, regType RegisterType) (int8, error) {
	b, err := r.rawRead(offset, 1, regType)
	if err != nil {
		return 0, err
	} else {
		return int8(b[0]), nil
	}
}

func (r *ViamModbusClientWithRetry) ReadUInt8(offset uint16, regType RegisterType) (uint8, error) {
	b, err := r.rawRead(offset, 1, regType)
	if err != nil {
		return 0, err
	} else {
		return uint8(b[0]), nil
	}
}

func (r *ViamModbusClientWithRetry) ReadInt16(offset uint16, regType RegisterType) (int16, error) {
	b, err := r.rawRead(offset, 1, regType)
	if err != nil {
		return 0, err
	} else {
		return int16(b[0]), nil
	}
}

func (r *ViamModbusClientWithRetry) ReadUInt16(offset uint16, regType RegisterType) (uint16, error) {
	b, err := r.rawRead(offset, 1, regType)
	if err != nil {
		return 0, err
	} else {
		return b[0], nil
	}
}

func (r *ViamModbusClientWithRetry) ReadInt32(offset uint16, regType RegisterType) (int32, error) {
	b, err := r.rawRead(offset, 2, regType)
	if err != nil {
		return 0, err
	} else {
		value := int32(b[0])<<16 | int32(b[1])
		return value, nil
	}
}

func (r *ViamModbusClientWithRetry) ReadUInt32(offset uint16, regType RegisterType) (uint32, error) {
	b, err := r.rawRead(offset, 2, regType)
	if err != nil {
		return 0, err
	} else {
		value := uint32(b[0])<<16 | uint32(b[1])
		return value, nil
	}
}

func (r *ViamModbusClientWithRetry) ReadInt64(offset uint16, regType RegisterType) (int64, error) {
	b, err := r.rawRead(offset, 4, regType)
	if err != nil {
		return 0, err
	} else {
		value := int64(b[0])<<48 | int64(b[1])<<32 | int64(b[2])<<16 | int64(b[3])
		return value, nil
	}
}

func (r *ViamModbusClientWithRetry) ReadUInt64(offset uint16, regType RegisterType) (uint64, error) {
	b, err := r.rawRead(offset, 4, regType)
	if err != nil {
		return 0, err
	} else {
		value := uint64(b[0])<<48 | uint64(b[1])<<32 | uint64(b[2])<<16 | uint64(b[3])
		return value, nil
	}
}

func (r *ViamModbusClientWithRetry) ReadFloat32(offset uint16, regType RegisterType) (float32, error) {
	b, err := r.rawRead(offset, 4, regType)
	if err != nil {
		return 0, err
	} else {
		value := uint32(b[0])<<16 | uint32(b[1])
		return math.Float32frombits(value), nil
	}
}

func (r *ViamModbusClientWithRetry) ReadFloat64(offset uint16, regType RegisterType) (float64, error) {
	b, err := r.rawRead(offset, 8, regType)
	if err != nil {
		return 0, err
	} else {
		value := uint64(b[0])<<48 | uint64(b[1])<<32 | uint64(b[2])<<16 | uint64(b[3])
		return math.Float64frombits(value), nil
	}
}

func (r *ViamModbusClientWithRetry) WriteInt8(offset uint16, value int8) error {
	return r.WriteUInt8(offset, uint8(value))
}

func (r *ViamModbusClientWithRetry) WriteUInt8(offset uint16, value uint8) error {
	regValue := uint16(value)
	return r.executeOperationWithRetry(func() error {
		return r.modbusClient.WriteMultipleRegisters(r.serverAddress, offset, []uint16{regValue})
	})
}

func (r *ViamModbusClientWithRetry) WriteInt16(offset uint16, value int16) error {
	return r.WriteUInt16(offset, uint16(value))
}

func (r *ViamModbusClientWithRetry) WriteUInt16(offset uint16, value uint16) error {
	return r.executeOperationWithRetry(func() error {
		return r.modbusClient.WriteMultipleRegisters(r.serverAddress, offset, []uint16{value})
	})
}

func (r *ViamModbusClientWithRetry) WriteInt32(offset uint16, value int32) error {
	return r.WriteUInt32(offset, uint32(value))
}

func (r *ViamModbusClientWithRetry) WriteUInt32(offset uint16, value uint32) error {
	regValues := []uint16{
		uint16(value >> 16),
		uint16(value & 0xFFFF),
	}
	return r.executeOperationWithRetry(func() error {
		return r.modbusClient.WriteMultipleRegisters(r.serverAddress, offset, regValues)
	})
}

func (r *ViamModbusClientWithRetry) WriteInt64(offset uint16, value int64) error {
	return r.WriteUInt64(offset, uint64(value))
}

func (r *ViamModbusClientWithRetry) WriteUInt64(offset uint16, value uint64) error {
	regValues := []uint16{
		uint16(value >> 48),
		uint16(value >> 32),
		uint16(value >> 16),
		uint16(value & 0xFFFF),
	}
	return r.executeOperationWithRetry(func() error {
		return r.modbusClient.WriteMultipleRegisters(r.serverAddress, offset, regValues)
	})
}

func (r *ViamModbusClientWithRetry) WriteFloat32(offset uint16, value float32) error {
	rawValue := math.Float32bits(value)
	regValue := []uint16{
		uint16(rawValue >> 16),
		uint16(rawValue & 0xFFFF),
	}
	return r.executeOperationWithRetry(func() error {
		return r.modbusClient.WriteMultipleRegisters(r.serverAddress, offset, regValue)
	})
}

func (r *ViamModbusClientWithRetry) WriteFloat64(offset uint16, value float64) error {
	rawValue := math.Float64bits(value)
	regValues := []uint16{
		uint16(rawValue >> 48),
		uint16(rawValue >> 32),
		uint16(rawValue >> 16),
		uint16(rawValue & 0xFFFF),
	}
	return r.executeOperationWithRetry(func() error {
		return r.modbusClient.WriteMultipleRegisters(r.serverAddress, offset, regValues)
	})
}

func (r *ViamModbusClientWithRetry) executeOperationWithRetry(o func() error) error {
	availableRetries := 3

	for availableRetries > 0 {
		err := o()
		if err != nil {
			availableRetries--
			err := r.reinitializeModbusClient()
			if err != nil {
				return err
			}
		} else {
			return nil
		}
	}
	return ErrRetriesExhausted
}
