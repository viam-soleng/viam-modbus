package common

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/simonvetter/modbus"
	"go.viam.com/rdk/logging"
)

type ModbusTcpClientCloudConfig struct {
	Url        string `json:"url"`
	Timeout    int    `json:"timeout_ms"`
	Endianness string `json:"endianness"`
	WordOrder  string `json:"word_order"`
}

func (cfg *ModbusTcpClientCloudConfig) Validate() error {
	if cfg.Url == "" {
		return errors.New("url is required")
	}
	if cfg.Timeout < 0 {
		return errors.New("timeout must be non-negative")
	}
	if cfg.Endianness != "big" && cfg.Endianness != "little" {
		return fmt.Errorf("endianness must be %v or %v", "big", "little")
	}
	if cfg.WordOrder != "high" && cfg.WordOrder != "low" {
		return fmt.Errorf("word_order must be %v or %v", "high", "low")
	}
	return nil
}

type ModbusTcpClient struct {
	mu           sync.RWMutex
	logger       logging.Logger
	uri          string
	timeout      time.Duration
	endianness   modbus.Endianness
	wordOrder    modbus.WordOrder
	modbusClient *modbus.ModbusClient
}

func NewModbusTcpClient(logger logging.Logger, uri string, timeout time.Duration, endianness modbus.Endianness, wordOrder modbus.WordOrder) (*ModbusTcpClient, error) {
	client := &ModbusTcpClient{
		logger:     logger,
		uri:        uri,
		timeout:    timeout,
		endianness: endianness,
		wordOrder:  wordOrder,
	}
	err := client.initializeModbusClient()
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (r *ModbusTcpClient) Close() error {
	if r.modbusClient != nil {
		r.modbusClient.Close()
	}
	return nil
}

func (r *ModbusTcpClient) reinitializeModbusClient() error {
	r.logger.Warnf("Re-initializing modbus client")
	return r.initializeModbusClient()
}

func (r *ModbusTcpClient) initializeModbusClient() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     r.uri,
		Timeout: r.timeout,
	})
	if err != nil {
		r.logger.Errorf("Failed to create modbus client: %#v", err)
		return err
	}
	client.SetEncoding(r.endianness, r.wordOrder)
	err = client.Open()
	if err != nil {
		r.logger.Errorf("Failed to open modbus client: %#v", err)
		return err
	}
	r.modbusClient = client
	return nil
}

func (r *ModbusTcpClient) ReadCoils(offset, length uint16) ([]bool, error) {
	availableRetries := 3

	for availableRetries > 0 {
		b, err := r.modbusClient.ReadCoils(offset, length)
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

func (r *ModbusTcpClient) ReadCoil(offset uint16) (bool, error) {
	availableRetries := 3

	for availableRetries > 0 {
		b, err := r.modbusClient.ReadCoil(offset)
		if err != nil {
			availableRetries--
			err := r.reinitializeModbusClient()
			if err != nil {
				return false, err
			}
		} else {
			return b, nil
		}
	}
	return false, ErrRetriesExhausted
}

func (r *ModbusTcpClient) ReadDiscreteInputs(offset, length uint16) ([]bool, error) {
	availableRetries := 3

	for availableRetries > 0 {
		b, err := r.modbusClient.ReadDiscreteInputs(offset, length)
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

func (r *ModbusTcpClient) ReadDiscreteInput(offset uint16) (bool, error) {
	availableRetries := 3

	for availableRetries > 0 {
		b, err := r.modbusClient.ReadDiscreteInput(offset)
		if err != nil {
			availableRetries--
			err := r.reinitializeModbusClient()
			if err != nil {
				return false, err
			}
		} else {
			return b, nil
		}
	}
	return false, ErrRetriesExhausted
}

func (r *ModbusTcpClient) WriteCoil(offset uint16, value bool) error {
	availableRetries := 3

	for availableRetries > 0 {
		err := r.modbusClient.WriteCoil(offset, value)
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

func (r *ModbusTcpClient) ReadHoldingRegisters(offset, length uint16) ([]uint16, error) {
	availableRetries := 3

	for availableRetries > 0 {
		b, err := r.modbusClient.ReadRegisters(offset, length, modbus.HOLDING_REGISTER)
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

func (r *ModbusTcpClient) ReadInputRegisters(offset, length uint16) ([]uint16, error) {
	availableRetries := 3

	for availableRetries > 0 {
		b, err := r.modbusClient.ReadRegisters(offset, length, modbus.INPUT_REGISTER)
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

func (r *ModbusTcpClient) ReadInt32(offset uint16, regType modbus.RegType) (int32, error) {
	availableRetries := 3

	for availableRetries > 0 {
		v, _ := r.modbusClient.ReadRegisters(offset, 2, regType)
		r.logger.Infof("ReadInt32: %#v", v)
		b, err := r.modbusClient.ReadUint32(offset, regType)
		if err != nil {
			availableRetries--
			err := r.reinitializeModbusClient()
			if err != nil {
				return 0, err
			}
		} else {
			return int32(b), nil
		}
	}
	return 0, ErrRetriesExhausted
}

func (r *ModbusTcpClient) ReadUInt32(offset uint16, regType modbus.RegType) (uint32, error) {
	availableRetries := 3

	for availableRetries > 0 {
		v, _ := r.modbusClient.ReadRegisters(offset, 2, regType)
		r.logger.Infof("ReadUInt32: %#v", v)
		b, err := r.modbusClient.ReadUint32(offset, regType)
		if err != nil {
			availableRetries--
			err := r.reinitializeModbusClient()
			if err != nil {
				return 0, err
			}
		} else {
			return b, nil
		}
	}
	return 0, ErrRetriesExhausted
}

func (r *ModbusTcpClient) ReadUInt64(offset uint16, regType modbus.RegType) (uint64, error) {
	availableRetries := 3

	for availableRetries > 0 {
		b, err := r.modbusClient.ReadUint64(offset, regType)
		if err != nil {
			availableRetries--
			err := r.reinitializeModbusClient()
			if err != nil {
				return 0, err
			}
		} else {
			return b, nil
		}
	}
	return 0, ErrRetriesExhausted
}

func (r *ModbusTcpClient) ReadFloat32(offset uint16, regType modbus.RegType) (float32, error) {
	availableRetries := 3

	for availableRetries > 0 {
		b, err := r.modbusClient.ReadFloat32(offset, regType)
		if err != nil {
			availableRetries--
			err := r.reinitializeModbusClient()
			if err != nil {
				return 0, err
			}
		} else {
			return b, nil
		}
	}
	return 0, ErrRetriesExhausted
}

func (r *ModbusTcpClient) ReadFloat64(offset uint16, regType modbus.RegType) (float64, error) {
	availableRetries := 3

	for availableRetries > 0 {
		b, err := r.modbusClient.ReadFloat64(offset, regType)
		if err != nil {
			availableRetries--
			err := r.reinitializeModbusClient()
			if err != nil {
				return 0, err
			}
		} else {
			return b, nil
		}
	}
	return 0, ErrRetriesExhausted
}

func (r *ModbusTcpClient) ReadUint8(offset uint16, regType modbus.RegType) (uint8, error) {
	availableRetries := 3

	for availableRetries > 0 {
		b, err := r.modbusClient.ReadRegister(offset, regType)
		if err != nil {
			availableRetries--
			err := r.reinitializeModbusClient()
			if err != nil {
				return 0, err
			}
		} else {
			return uint8(b), nil
		}
	}
	return 0, ErrRetriesExhausted
}

func (r *ModbusTcpClient) ReadInt16(offset uint16, regType modbus.RegType) (int16, error) {
	availableRetries := 3

	for availableRetries > 0 {
		b, err := r.modbusClient.ReadRegister(offset, regType)
		if err != nil {
			availableRetries--
			err := r.reinitializeModbusClient()
			if err != nil {
				return 0, err
			}
		} else {
			return int16(b), nil
		}
	}
	return 0, ErrRetriesExhausted
}
func (r *ModbusTcpClient) ReadUInt16(offset uint16, regType modbus.RegType) (uint16, error) {
	availableRetries := 3

	for availableRetries > 0 {
		b, err := r.modbusClient.ReadRegister(offset, regType)
		if err != nil {
			availableRetries--
			err := r.reinitializeModbusClient()
			if err != nil {
				return 0, err
			}
		} else {
			return b, nil
		}
	}
	return 0, ErrRetriesExhausted
}

func (r *ModbusTcpClient) ReadBytes(offset, length uint16, regType modbus.RegType) ([]byte, error) {
	availableRetries := 3

	for availableRetries > 0 {
		b, err := r.modbusClient.ReadBytes(offset, length, regType)
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

func (r *ModbusTcpClient) ReadRawBytes(offset, length uint16, regType modbus.RegType) ([]byte, error) {
	availableRetries := 3

	for availableRetries > 0 {
		b, err := r.modbusClient.ReadRawBytes(offset, length, regType)
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
