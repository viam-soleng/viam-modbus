package common

import (
	"errors"
	"fmt"
	"sync"

	"github.com/simonvetter/modbus"
	"go.viam.com/rdk/logging"
)

type ModbusClientConfig struct {
	URL           string `json:"url"`
	Speed         uint   `json:"speed"`
	DataBits      uint   `json:"data_bits"`
	Parity        uint   `json:"parity"`
	StopBits      uint   `json:"stop_bits"`
	Timeout       int    `json:"timeout_ms"`
	TLSClientCert string `json:"tls_client_cert"`
	TLSRootCAs    string `json:"tls_root_cas"`
	Endianness    string `json:"endianness"`
	WordOrder     string `json:"word_order"`
}

func (cfg *ModbusClientConfig) Validate() error {
	if cfg.URL == "" {
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

type ViamModbusClient struct {
	mu     sync.RWMutex
	logger logging.Logger

	endianness   modbus.Endianness
	wordOrder    modbus.WordOrder
	modbusClient *modbus.ModbusClient
	modbus.ClientConfiguration
}

// TODO: Need to make it so reconfigure can update the settings here when stuck in a re-retry loop
func NewModbusClient(logger logging.Logger, endianness modbus.Endianness, wordOrder modbus.WordOrder, clientConfig modbus.ClientConfiguration) (*ViamModbusClient, error) {
	client := &ViamModbusClient{
		logger:              logger,
		endianness:          endianness,
		wordOrder:           wordOrder,
		ClientConfiguration: clientConfig,
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

	client, err := modbus.NewClient(&r.ClientConfiguration)
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

func (r *ViamModbusClient) ReadCoils(offset, length uint16, unitID int8) ([]bool, error) {
	availableRetries := 3

	for availableRetries > 0 {
		if unitID >= 0 {
			r.modbusClient.SetUnitId(uint8(unitID))
		}
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

func (r *ViamModbusClient) ReadCoil(offset uint16, unitID int8) (bool, error) {
	availableRetries := 3

	for availableRetries > 0 {
		if unitID >= 0 {
			r.modbusClient.SetUnitId(uint8(unitID))
		}
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

func (r *ViamModbusClient) ReadDiscreteInputs(offset, length uint16, unitID int8) ([]bool, error) {
	availableRetries := 3

	for availableRetries > 0 {
		if unitID >= 0 {
			r.modbusClient.SetUnitId(uint8(unitID))
		}
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

func (r *ViamModbusClient) ReadDiscreteInput(offset uint16, unitID int8) (bool, error) {
	availableRetries := 3

	for availableRetries > 0 {
		if unitID >= 0 {
			r.modbusClient.SetUnitId(uint8(unitID))
		}
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

func (r *ViamModbusClient) WriteCoil(offset uint16, value bool, unitID int8) error {
	availableRetries := 3

	for availableRetries > 0 {
		if unitID >= 0 {
			r.modbusClient.SetUnitId(uint8(unitID))
		}
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

func (r *ViamModbusClient) ReadHoldingRegisters(offset, length uint16, unitID int8) ([]uint16, error) {
	availableRetries := 3

	for availableRetries > 0 {
		if unitID >= 0 {
			r.modbusClient.SetUnitId(uint8(unitID))
		}
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

func (r *ViamModbusClient) ReadInputRegisters(offset, length uint16, unitID int8) ([]uint16, error) {
	availableRetries := 3

	for availableRetries > 0 {
		if unitID >= 0 {
			r.modbusClient.SetUnitId(uint8(unitID))
		}
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

func (r *ViamModbusClient) ReadInt32(offset uint16, regType modbus.RegType, unitID int8) (int32, error) {
	availableRetries := 3

	for availableRetries > 0 {
		if unitID >= 0 {
			r.modbusClient.SetUnitId(uint8(unitID))
		}
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

func (r *ViamModbusClient) ReadUInt32(offset uint16, regType modbus.RegType, unitID int8) (uint32, error) {
	availableRetries := 3

	for availableRetries > 0 {
		if unitID >= 0 {
			r.modbusClient.SetUnitId(uint8(unitID))
		}
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

func (r *ViamModbusClient) ReadUInt64(offset uint16, regType modbus.RegType, unitID int8) (uint64, error) {
	availableRetries := 3

	for availableRetries > 0 {
		if unitID >= 0 {
			r.modbusClient.SetUnitId(uint8(unitID))
		}
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

func (r *ViamModbusClient) ReadFloat32(offset uint16, regType modbus.RegType, unitID int8) (float32, error) {
	availableRetries := 3

	for availableRetries > 0 {
		if unitID >= 0 {
			r.modbusClient.SetUnitId(uint8(unitID))
		}
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

func (r *ViamModbusClient) ReadFloat64(offset uint16, regType modbus.RegType, unitID int8) (float64, error) {
	availableRetries := 3

	for availableRetries > 0 {
		if unitID >= 0 {
			r.modbusClient.SetUnitId(uint8(unitID))
		}
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

func (r *ViamModbusClient) ReadUInt8(offset uint16, regType modbus.RegType, unitID int8) (uint8, error) {
	availableRetries := 3

	for availableRetries > 0 {
		if unitID >= 0 {
			r.modbusClient.SetUnitId(uint8(unitID))
		}
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

func (r *ViamModbusClient) ReadInt16(offset uint16, regType modbus.RegType, unitID int8) (int16, error) {
	availableRetries := 3

	for availableRetries > 0 {
		if unitID >= 0 {
			r.modbusClient.SetUnitId(uint8(unitID))
		}
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

func (r *ViamModbusClient) ReadUInt16(offset uint16, regType modbus.RegType, unitID int8) (uint16, error) {
	availableRetries := 3

	for availableRetries > 0 {
		if unitID >= 0 {
			r.modbusClient.SetUnitId(uint8(unitID))
		}
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

func (r *ViamModbusClient) ReadBytes(offset, length uint16, regType modbus.RegType, unitID int8) ([]byte, error) {
	availableRetries := 3

	for availableRetries > 0 {
		if unitID >= 0 {
			r.modbusClient.SetUnitId(uint8(unitID))
		}
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

func (r *ViamModbusClient) ReadRawBytes(offset, length uint16, regType modbus.RegType, unitID int8) ([]byte, error) {
	availableRetries := 3

	for availableRetries > 0 {
		if unitID >= 0 {
			r.modbusClient.SetUnitId(uint8(unitID))
		}
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

func (r *ViamModbusClient) WriteUInt16(offset uint16, value uint16, unitID int8) error {
	return r.WriteWithRetry(func() error {
		return r.modbusClient.WriteRegister(offset, value)
	}, unitID)
}

func (r *ViamModbusClient) WriteUInt32(offset uint16, value uint32, unitID int8) error {
	return r.WriteWithRetry(func() error {
		return r.modbusClient.WriteUint32(offset, value)
	}, unitID)
}

func (r *ViamModbusClient) WriteUInt64(offset uint16, value uint64, unitID int8) error {
	return r.WriteWithRetry(func() error {
		return r.modbusClient.WriteUint64(offset, value)
	}, unitID)
}

func (r *ViamModbusClient) WriteFloat32(offset uint16, value float32, unitID int8) error {
	return r.WriteWithRetry(func() error {
		return r.modbusClient.WriteFloat32(offset, value)
	}, unitID)
}

func (r *ViamModbusClient) WriteFloat64(offset uint16, value float64, unitID int8) error {
	return r.WriteWithRetry(func() error {
		return r.modbusClient.WriteFloat64(offset, value)
	}, unitID)
}

func (r *ViamModbusClient) WriteWithRetry(w func() error, unitID int8) error {
	availableRetries := 3

	for availableRetries > 0 {
		if unitID >= 0 {
			r.modbusClient.SetUnitId(uint8(unitID))
		}
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
