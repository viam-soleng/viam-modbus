package common

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/simonvetter/modbus"
	"go.viam.com/rdk/logging"
)

type ModbusClientCloudConfig struct {
	Url        string `json:"url"`
	Speed      uint   `json:"speed"`
	Timeout    int    `json:"timeout_ms"`
	Endianness string `json:"endianness"`
	WordOrder  string `json:"word_order"`
}

func (cfg *ModbusClientCloudConfig) Validate() error {
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

type ModbusClient struct {
	mu           sync.RWMutex
	logger       logging.Logger
	uri          string
	speed        uint
	timeout      time.Duration
	endianness   modbus.Endianness
	wordOrder    modbus.WordOrder
	modbusClient *modbus.ModbusClient
}

// TODO: Need to make it so reconfigure can update the settings here when stuck in a re-retry loop
func NewModbusTcpClient(logger logging.Logger, uri string, timeout time.Duration, endianness modbus.Endianness, wordOrder modbus.WordOrder, speed uint) (*ModbusClient, error) {
	client := &ModbusClient{
		logger:     logger,
		uri:        uri,
		speed:      speed,
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

func (r *ModbusClient) Close() error {
	if r.modbusClient != nil {
		r.modbusClient.Close()
	}
	return nil
}

func (r *ModbusClient) reinitializeModbusClient() error {
	r.logger.Warnf("Re-initializing modbus client")
	return r.initializeModbusClient()
}

func (r *ModbusClient) initializeModbusClient() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// TODO: for an RTU (serial) device/bus
	/*
	   client, err = modbus.NewClient(&modbus.ClientConfiguration{
	       URL:      "rtu:///dev/ttyUSB0",
	       Speed:    19200,                   // default
	       DataBits: 8,                       // default, optional
	       Parity:   modbus.PARITY_NONE,      // default, optional
	       StopBits: 2,                       // default if no parity, optional
	       Timeout:  300 * time.Millisecond,
	   })
	*/

	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     r.uri,
		Speed:   r.speed,
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

func (r *ModbusClient) ReadCoils(offset, length uint16) ([]bool, error) {
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

func (r *ModbusClient) ReadCoil(offset uint16) (bool, error) {
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

func (r *ModbusClient) ReadDiscreteInputs(offset, length uint16) ([]bool, error) {
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

func (r *ModbusClient) ReadDiscreteInput(offset uint16) (bool, error) {
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

func (r *ModbusClient) WriteCoil(offset uint16, value bool) error {
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

func (r *ModbusClient) ReadHoldingRegisters(offset, length uint16) ([]uint16, error) {
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

func (r *ModbusClient) ReadInputRegisters(offset, length uint16) ([]uint16, error) {
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

func (r *ModbusClient) ReadInt32(offset uint16, regType modbus.RegType) (int32, error) {
	availableRetries := 3

	for availableRetries > 0 {
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

func (r *ModbusClient) ReadUInt32(offset uint16, regType modbus.RegType) (uint32, error) {
	availableRetries := 3

	for availableRetries > 0 {
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

func (r *ModbusClient) ReadUInt64(offset uint16, regType modbus.RegType) (uint64, error) {
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

func (r *ModbusClient) ReadFloat32(offset uint16, regType modbus.RegType) (float32, error) {
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

func (r *ModbusClient) ReadFloat64(offset uint16, regType modbus.RegType) (float64, error) {
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

func (r *ModbusClient) ReadUInt8(offset uint16, regType modbus.RegType) (uint8, error) {
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

func (r *ModbusClient) ReadInt16(offset uint16, regType modbus.RegType) (int16, error) {
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

func (r *ModbusClient) ReadUInt16(offset uint16, regType modbus.RegType) (uint16, error) {
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

func (r *ModbusClient) ReadBytes(offset, length uint16, regType modbus.RegType) ([]byte, error) {
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

func (r *ModbusClient) ReadRawBytes(offset, length uint16, regType modbus.RegType) ([]byte, error) {
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

func (r *ModbusClient) WriteUInt16(offset uint16, value uint16) error {
	return r.WriteWithRetry(func() error {
		return r.modbusClient.WriteRegister(offset, value)
	})
}

func (r *ModbusClient) WriteUInt32(offset uint16, value uint32) error {
	return r.WriteWithRetry(func() error {
		return r.modbusClient.WriteUint32(offset, value)
	})
}

func (r *ModbusClient) WriteUInt64(offset uint16, value uint64) error {
	return r.WriteWithRetry(func() error {
		return r.modbusClient.WriteUint64(offset, value)
	})
}

func (r *ModbusClient) WriteFloat32(offset uint16, value float32) error {
	return r.WriteWithRetry(func() error {
		return r.modbusClient.WriteFloat32(offset, value)
	})
}

func (r *ModbusClient) WriteFloat64(offset uint16, value float64) error {
	return r.WriteWithRetry(func() error {
		return r.modbusClient.WriteFloat64(offset, value)
	})
}

func (r *ModbusClient) WriteWithRetry(w func() error) error {
	availableRetries := 3

	for availableRetries > 0 {
		err := w()
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
