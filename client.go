package viammodbus

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/simonvetter/modbus"
	"go.viam.com/rdk/components/generic"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
)

var NamespaceFamily = resource.NewModelFamily("viam-soleng", "modbus")
var ModbusClientModel = NamespaceFamily.WithModel("client")

var ErrRetriesExhausted = fmt.Errorf("retries exhausted")

func init() {
	resource.RegisterComponent(
		generic.API,
		ModbusClientModel,
		resource.Registration[generic.Resource, *modbusClientConfig]{
			Constructor: newModbusClient,
		})
}

type modbusClientConfig struct {
	URL           string `json:"url"`
	Speed         uint   `json:"speed"`
	DataBits      uint   `json:"data_bits"`
	Parity        uint   `json:"parity"`
	StopBits      uint   `json:"stop_bits"`
	Timeout       int    `json:"timeout_ms"`
	Endianness    string `json:"endianness"`
	WordOrder     string `json:"word_order"`
	TLSClientCert string `json:"tls_client_cert"`
	TLSRootCAs    string `json:"tls_root_cas"`
}

func (cfg *modbusClientConfig) Validate(path string) ([]string, []string, error) {
	if cfg.URL == "" {
		return nil, nil, fmt.Errorf("url is required")
	}
	if cfg.Timeout < 0 {
		return nil, nil, fmt.Errorf("timeout must be non-negative")
	}
	if cfg.Endianness != "big" && cfg.Endianness != "little" {
		return nil, nil, fmt.Errorf("endianness must be %v or %v", "big", "little")
	}
	if cfg.WordOrder != "high" && cfg.WordOrder != "low" {
		return nil, nil, fmt.Errorf("word_order must be %v or %v", "high", "low")
	}
	if cfg.TLSClientCert != "" || cfg.TLSRootCAs != "" {
		fmt.Println("Warning: TLS is not supported yet, TLSClientCert and TLSRootCAs will be ignored")
	}
	return []string{}, nil, nil
}

type modbusClient struct {
	resource.AlwaysRebuild
	name   resource.Name
	mu     sync.RWMutex
	logger logging.Logger

	endianness modbus.Endianness
	wordOrder  modbus.WordOrder

	client *modbus.ModbusClient
	config modbus.ClientConfiguration
}

func newModbusClient(ctx context.Context, deps resource.Dependencies, config resource.Config, logger logging.Logger) (generic.Resource, error) {
	newConf, err := resource.NativeConfig[*modbusClientConfig](config)
	if err != nil {
		return nil, err
	}

	endianness, err := GetEndianness(newConf.Endianness)
	if err != nil {
		return nil, err
	}

	wordOrder, err := GetWordOrder(newConf.WordOrder)
	if err != nil {
		return nil, err
	}

	timeout := time.Millisecond * time.Duration(newConf.Timeout)

	clientConfig := modbus.ClientConfiguration{
		URL:      newConf.URL,
		Speed:    newConf.Speed,
		DataBits: newConf.DataBits,
		Parity:   newConf.Parity,
		StopBits: newConf.StopBits,
		Timeout:  timeout,
		//TODO: Add TLS support
		//TLSClientCert: newConf.TLSClientCert,
		//TLSRootCAs:    newConf.TLSRootCAs,
	}

	client := &modbusClient{
		name:       config.ResourceName(),
		logger:     logger,
		endianness: endianness,
		wordOrder:  wordOrder,
		config:     clientConfig,
	}

	err = GlobalClientRegistry.Add(client.name.Name, client)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (gs *modbusClient) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("DoCommand not implemented")
}

func (gs *modbusClient) Close(ctx context.Context) error {
	GlobalClientRegistry.Remove(gs.name.Name)
	return nil
}

func (gs *modbusClient) Name() resource.Name {
	return gs.name
}

func (r *modbusClient) ReadCoils(offset, length uint16, unitID *uint8) ([]bool, error) {
	availableRetries := 3

	for availableRetries > 0 {
		if unitID != nil {
			r.client.SetUnitId(uint8(*unitID))
		}
		b, err := r.client.ReadCoils(offset, length)
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

func (r *modbusClient) ReadCoil(offset uint16, unitID *uint8) (bool, error) {
	availableRetries := 3

	for availableRetries > 0 {
		if unitID != nil {
			r.client.SetUnitId(*unitID)
		}
		b, err := r.client.ReadCoil(offset)
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

func (r *modbusClient) ReadDiscreteInputs(offset, length uint16, unitID *uint8) ([]bool, error) {
	availableRetries := 3

	for availableRetries > 0 {
		if unitID != nil {
			r.client.SetUnitId(*unitID)
		}
		b, err := r.client.ReadDiscreteInputs(offset, length)
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

func (r *modbusClient) ReadDiscreteInput(offset uint16, unitID *uint8) (bool, error) {
	availableRetries := 3

	for availableRetries > 0 {
		if unitID != nil {
			r.client.SetUnitId(*unitID)
		}
		b, err := r.client.ReadDiscreteInput(offset)
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

func (r *modbusClient) WriteCoil(offset uint16, value bool, unitID *uint8) error {
	availableRetries := 3

	for availableRetries > 0 {
		if unitID != nil {
			r.client.SetUnitId(*unitID)
		}
		err := r.client.WriteCoil(offset, value)
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

func (r *modbusClient) ReadHoldingRegisters(offset, length uint16, unitID *uint8) ([]uint16, error) {
	availableRetries := 3

	for availableRetries > 0 {
		if unitID != nil {
			r.client.SetUnitId(*unitID)
		}
		b, err := r.client.ReadRegisters(offset, length, modbus.HOLDING_REGISTER)
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

func (r *modbusClient) ReadInputRegisters(offset, length uint16, unitID *uint8) ([]uint16, error) {
	availableRetries := 3

	for availableRetries > 0 {
		if unitID != nil {
			r.client.SetUnitId(*unitID)
		}
		b, err := r.client.ReadRegisters(offset, length, modbus.INPUT_REGISTER)
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

func (r *modbusClient) ReadInt32(offset uint16, regType modbus.RegType, unitID *uint8) (int32, error) {
	availableRetries := 3

	for availableRetries > 0 {
		if unitID != nil {
			r.client.SetUnitId(*unitID)
		}
		b, err := r.client.ReadUint32(offset, regType)
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

func (r *modbusClient) ReadUInt32(offset uint16, regType modbus.RegType, unitID *uint8) (uint32, error) {
	availableRetries := 3

	for availableRetries > 0 {
		if unitID != nil {
			r.client.SetUnitId(*unitID)
		}
		b, err := r.client.ReadUint32(offset, regType)
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

func (r *modbusClient) ReadUInt64(offset uint16, regType modbus.RegType, unitID *uint8) (uint64, error) {
	availableRetries := 3

	for availableRetries > 0 {
		if unitID != nil {
			r.client.SetUnitId(*unitID)
		}
		b, err := r.client.ReadUint64(offset, regType)
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

func (r *modbusClient) ReadFloat32(offset uint16, regType modbus.RegType, unitID *uint8) (float32, error) {
	availableRetries := 3

	for availableRetries > 0 {
		if unitID != nil {
			r.client.SetUnitId(*unitID)
		}
		b, err := r.client.ReadFloat32(offset, regType)
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

func (r *modbusClient) ReadFloat64(offset uint16, regType modbus.RegType, unitID *uint8) (float64, error) {
	availableRetries := 3

	for availableRetries > 0 {
		if unitID != nil {
			r.client.SetUnitId(*unitID)
		}
		b, err := r.client.ReadFloat64(offset, regType)
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

func (r *modbusClient) ReadUInt8(offset uint16, regType modbus.RegType, unitID *uint8) (uint8, error) {
	availableRetries := 3

	for availableRetries > 0 {
		if unitID != nil {
			r.client.SetUnitId(*unitID)
		}
		b, err := r.client.ReadRegister(offset, regType)
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

func (r *modbusClient) ReadInt16(offset uint16, regType modbus.RegType, unitID *uint8) (int16, error) {
	availableRetries := 3

	for availableRetries > 0 {
		if unitID != nil {
			r.client.SetUnitId(*unitID)
		}
		b, err := r.client.ReadRegister(offset, regType)
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

func (r *modbusClient) ReadUInt16(offset uint16, regType modbus.RegType, unitID *uint8) (uint16, error) {
	availableRetries := 3

	for availableRetries > 0 {
		if unitID != nil {
			r.client.SetUnitId(*unitID)
		}
		b, err := r.client.ReadRegister(offset, regType)
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

func (r *modbusClient) ReadBytes(offset, length uint16, regType modbus.RegType, unitID *uint8) ([]byte, error) {
	availableRetries := 3

	for availableRetries > 0 {
		if unitID != nil {
			r.client.SetUnitId(*unitID)
		}
		b, err := r.client.ReadBytes(offset, length, regType)
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

func (r *modbusClient) ReadRawBytes(offset, length uint16, regType modbus.RegType, unitID *uint8) ([]byte, error) {
	availableRetries := 3

	for availableRetries > 0 {
		if unitID != nil {
			r.client.SetUnitId(*unitID)
		}
		b, err := r.client.ReadRawBytes(offset, length, regType)
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

func (r *modbusClient) WriteUInt16(offset uint16, value uint16, unitID *uint8) error {
	return r.WriteWithRetry(func() error {
		return r.client.WriteRegister(offset, value)
	}, unitID)
}

func (r *modbusClient) WriteUInt32(offset uint16, value uint32, unitID *uint8) error {
	return r.WriteWithRetry(func() error {
		return r.client.WriteUint32(offset, value)
	}, unitID)
}

func (r *modbusClient) WriteUInt64(offset uint16, value uint64, unitID *uint8) error {
	return r.WriteWithRetry(func() error {
		return r.client.WriteUint64(offset, value)
	}, unitID)
}

func (r *modbusClient) WriteFloat32(offset uint16, value float32, unitID *uint8) error {
	return r.WriteWithRetry(func() error {
		return r.client.WriteFloat32(offset, value)
	}, unitID)
}

func (r *modbusClient) WriteFloat64(offset uint16, value float64, unitID *uint8) error {
	return r.WriteWithRetry(func() error {
		return r.client.WriteFloat64(offset, value)
	}, unitID)
}

func (r *modbusClient) WriteWithRetry(w func() error, unitID *uint8) error {
	availableRetries := 3

	for availableRetries > 0 {
		if unitID != nil {
			r.client.SetUnitId(*unitID)
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

func (r *modbusClient) initializeModbusClient() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	//TODO: Don't think this is needed or shall be pushed to the client registry
	r.client.Open()

	return nil
}

func (r *modbusClient) reinitializeModbusClient() error {
	r.logger.Warnf("Re-initializing modbus client")
	return r.initializeModbusClient()
}

func GetEndianness(s string) (modbus.Endianness, error) {
	switch s {
	case "big":
		return modbus.BIG_ENDIAN, nil
	case "little":
		return modbus.LITTLE_ENDIAN, nil
	default:
		return 0, fmt.Errorf("invalid endianness")
	}
}

func GetWordOrder(s string) (modbus.WordOrder, error) {
	switch s {
	case "high":
		return modbus.HIGH_WORD_FIRST, nil
	case "low":
		return modbus.LOW_WORD_FIRST, nil
	default:
		return 0, fmt.Errorf("invalid word order")
	}
}
