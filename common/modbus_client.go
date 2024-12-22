// Package common provides common functionality across all modbus devices in this module
package common

import (
	"errors"
	"math"
	"sync"
	"time"

	"github.com/goburrow/serial"
	"github.com/rinzlerlabs/gomodbus/client"
	"github.com/rinzlerlabs/gomodbus/client/network/tcp"
	"github.com/rinzlerlabs/gomodbus/client/serial/ascii"
	"github.com/rinzlerlabs/gomodbus/client/serial/rtu"
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

type ViamModbusClient struct {
	mu            sync.RWMutex
	logger        logging.Logger
	conf          *ModbusConfig
	modbusClient  client.ModbusClient
	serverAddress uint16
}

func NewModbusClient(logger logging.Logger, conf *ModbusConfig) (*ViamModbusClient, error) {
	client := &ViamModbusClient{
		logger: logger,
	}
	err := client.initializeModbusClient()
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (r *ViamModbusClient) Close() error {
	if r.modbusClient != nil {
		r.modbusClient.Close()
	}
	return nil
}

func (r *ViamModbusClient) reinitializeModbusClient() error {
	r.logger.Warnf("Re-initializing modbus client")
	return r.initializeModbusClient()
}

func (r *ViamModbusClient) initializeModbusClient() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.modbusClient != nil {
		_ = r.modbusClient.Close()
	}

	var client client.ModbusClient
	var err error
	if r.conf.SerialConfig != nil {
		serialConfig := &serial.Config{
			Address:  r.conf.Endpoint,
			BaudRate: int(r.conf.SerialConfig.Speed),
			DataBits: int(r.conf.SerialConfig.DataBits),
			StopBits: int(r.conf.SerialConfig.StopBits),
			Parity:   r.conf.SerialConfig.Parity,
		}
		port, err := serial.Open(serialConfig)
		if err != nil {
			return err
		}
		if r.conf.SerialConfig.RTU {
			client = rtu.NewModbusClient(r.logger.AsZap().Desugar(), port, 1*time.Second)
		} else {
			client = ascii.NewModbusClient(r.logger.AsZap().Desugar(), port, 1*time.Second)
		}
	} else if r.conf.TCPConfig != nil {
		client, err = tcp.NewModbusClient(r.logger.AsZap().Desugar(), r.conf.Endpoint, 1*time.Second)
	} else {
		return errors.New("invalid config")
	}
	if err != nil {
		r.logger.Errorf("Failed to initialize modbus client: %#v", err)
		return err
	}
	r.modbusClient = client
	return nil
}

func (r *ViamModbusClient) ReadCoils(offset, length uint16) ([]bool, error) {
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

func (r *ViamModbusClient) ReadCoil(offset uint16) (bool, error) {
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

func (r *ViamModbusClient) ReadDiscreteInputs(offset, length uint16) ([]bool, error) {
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

func (r *ViamModbusClient) ReadDiscreteInput(offset uint16) (bool, error) {
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

func (r *ViamModbusClient) WriteCoil(offset uint16, value bool) error {
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

func (r *ViamModbusClient) ReadHoldingRegisters(offset, length uint16) ([]uint16, error) {
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

func (r *ViamModbusClient) ReadInputRegisters(offset, length uint16) ([]uint16, error) {
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
func (r *ViamModbusClient) rawRead(offset, len uint16, regType RegisterType) ([]uint16, error) {
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

func (r *ViamModbusClient) ReadInt8(offset uint16, regType RegisterType) (int8, error) {
	b, err := r.rawRead(offset, 1, regType)
	if err != nil {
		return 0, err
	} else {
		return int8(b[0]), nil
	}
}

func (r *ViamModbusClient) ReadUInt8(offset uint16, regType RegisterType) (uint8, error) {
	b, err := r.rawRead(offset, 1, regType)
	if err != nil {
		return 0, err
	} else {
		return uint8(b[0]), nil
	}
}

func (r *ViamModbusClient) ReadInt16(offset uint16, regType RegisterType) (int16, error) {
	b, err := r.rawRead(offset, 1, regType)
	if err != nil {
		return 0, err
	} else {
		return int16(b[0]), nil
	}
}

func (r *ViamModbusClient) ReadUInt16(offset uint16, regType RegisterType) (uint16, error) {
	b, err := r.rawRead(offset, 1, regType)
	if err != nil {
		return 0, err
	} else {
		return b[0], nil
	}
}

func (r *ViamModbusClient) ReadInt32(offset uint16, regType RegisterType) (int32, error) {
	b, err := r.rawRead(offset, 2, regType)
	if err != nil {
		return 0, err
	} else {
		value := int32(b[0])<<16 | int32(b[1])
		return value, nil
	}
}

func (r *ViamModbusClient) ReadUInt32(offset uint16, regType RegisterType) (uint32, error) {
	b, err := r.rawRead(offset, 2, regType)
	if err != nil {
		return 0, err
	} else {
		value := uint32(b[0])<<16 | uint32(b[1])
		return value, nil
	}
}

func (r *ViamModbusClient) ReadInt64(offset uint16, regType RegisterType) (int64, error) {
	b, err := r.rawRead(offset, 4, regType)
	if err != nil {
		return 0, err
	} else {
		value := int64(b[0])<<48 | int64(b[1])<<32 | int64(b[2])<<16 | int64(b[3])
		return value, nil
	}
}

func (r *ViamModbusClient) ReadUInt64(offset uint16, regType RegisterType) (uint64, error) {
	b, err := r.rawRead(offset, 4, regType)
	if err != nil {
		return 0, err
	} else {
		value := uint64(b[0])<<48 | uint64(b[1])<<32 | uint64(b[2])<<16 | uint64(b[3])
		return value, nil
	}
}

func (r *ViamModbusClient) ReadFloat32(offset uint16, regType RegisterType) (float32, error) {
	b, err := r.rawRead(offset, 4, regType)
	if err != nil {
		return 0, err
	} else {
		value := uint32(b[0])<<16 | uint32(b[1])
		return math.Float32frombits(value), nil
	}
}

func (r *ViamModbusClient) ReadFloat64(offset uint16, regType RegisterType) (float64, error) {
	b, err := r.rawRead(offset, 8, regType)
	if err != nil {
		return 0, err
	} else {
		value := uint64(b[0])<<48 | uint64(b[1])<<32 | uint64(b[2])<<16 | uint64(b[3])
		return math.Float64frombits(value), nil
	}
}

func (r *ViamModbusClient) WriteInt8(offset uint16, value int8) error {
	return r.WriteUInt8(offset, uint8(value))
}

func (r *ViamModbusClient) WriteUInt8(offset uint16, value uint8) error {
	regValue := uint16(value)
	return r.executeOperationWithRetry(func() error {
		return r.modbusClient.WriteMultipleRegisters(r.serverAddress, offset, []uint16{regValue})
	})
}

func (r *ViamModbusClient) WriteInt16(offset uint16, value int16) error {
	return r.WriteUInt16(offset, uint16(value))
}

func (r *ViamModbusClient) WriteUInt16(offset uint16, value uint16) error {
	return r.executeOperationWithRetry(func() error {
		return r.modbusClient.WriteMultipleRegisters(r.serverAddress, offset, []uint16{value})
	})
}

func (r *ViamModbusClient) WriteInt32(offset uint16, value int32) error {
	return r.WriteUInt32(offset, uint32(value))
}

func (r *ViamModbusClient) WriteUInt32(offset uint16, value uint32) error {
	regValues := []uint16{
		uint16(value >> 16),
		uint16(value & 0xFFFF),
	}
	return r.executeOperationWithRetry(func() error {
		return r.modbusClient.WriteMultipleRegisters(r.serverAddress, offset, regValues)
	})
}

func (r *ViamModbusClient) WriteInt64(offset uint16, value int64) error {
	return r.WriteUInt64(offset, uint64(value))
}

func (r *ViamModbusClient) WriteUInt64(offset uint16, value uint64) error {
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

func (r *ViamModbusClient) WriteFloat32(offset uint16, value float32) error {
	rawValue := math.Float32bits(value)
	regValue := []uint16{
		uint16(rawValue >> 16),
		uint16(rawValue & 0xFFFF),
	}
	return r.executeOperationWithRetry(func() error {
		return r.modbusClient.WriteMultipleRegisters(r.serverAddress, offset, regValue)
	})
}

func (r *ViamModbusClient) WriteFloat64(offset uint16, value float64) error {
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

func (r *ViamModbusClient) executeOperationWithRetry(o func() error) error {
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
