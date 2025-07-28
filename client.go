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
	if cfg.Endianness != "" && cfg.Endianness != "big" && cfg.Endianness != "little" {
		return nil, nil, fmt.Errorf("endianness must be %v or %v", "big", "little")
	}
	if cfg.WordOrder != "" && cfg.WordOrder != "high" && cfg.WordOrder != "low" {
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

	endianness *modbus.Endianness
	wordOrder  *modbus.WordOrder

	client *modbus.ModbusClient
	config modbus.ClientConfiguration
}

func newModbusClient(ctx context.Context, deps resource.Dependencies, config resource.Config, logger logging.Logger) (generic.Resource, error) {
	newConf, err := resource.NativeConfig[*modbusClientConfig](config)
	if err != nil {
		return nil, err
	}

	var timeout time.Duration
	if newConf.Timeout > 0 {
		timeout = time.Millisecond * time.Duration(newConf.Timeout)
	}

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
		name:   config.ResourceName(),
		logger: logger,
		config: clientConfig,
	}

	// Create the modbus connection with the provided configuration
	err = client.newModbusConnection(&clientConfig)
	if err != nil {
		logger.Errorf("Failed to create modbus client: %#v", err)
		return nil, err
	}

	// Set the endianness and word order for the client
	setDecoding(client, newConf.Endianness, newConf.WordOrder)

	// Add the modbus client to the registry
	err = GlobalClientRegistry.Add(client.name.Name, client)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (mc *modbusClient) newModbusConnection(config *modbus.ClientConfiguration) error {
	mc.logger.Infof("Creating new modbus client with config: %#v", config)
	client, err := modbus.NewClient(config)
	if err != nil {
		return err
	}
	mc.client = client
	// now that the client is created and configured, attempt to connect
	err = mc.client.Open()
	if err != nil {
		mc.logger.Errorf("Failed to open modbus client: %#v", err)
		return err
	}
	return nil
}

func setDecoding(mc *modbusClient, endianness, wordOrder string) {
	if endianness != "" {
		enc, err := GetEndianness(endianness)
		if err != nil {
			mc.logger.Errorf("Invalid endianness: %v", err)
			return
		}
		mc.endianness = &enc
	}
	if wordOrder != "" {
		wo, err := GetWordOrder(wordOrder)
		if err != nil {
			mc.logger.Errorf("Invalid word order: %v", err)
			return
		}
		mc.wordOrder = &wo
	}
	if mc.endianness != nil && mc.wordOrder != nil {
		mc.client.SetEncoding(*mc.endianness, *mc.wordOrder)
		mc.logger.Infof("Set endianness to %v and word order to %v", endianness, wordOrder)
	} else {
		mc.logger.Infof("Using default endianness and word order")
	}
}

func (mc *modbusClient) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("DoCommand not implemented")
}

func (mc *modbusClient) Close(ctx context.Context) error {
	GlobalClientRegistry.Remove(mc.name.Name)
	return nil
}

func (mc *modbusClient) Name() resource.Name {
	return mc.name
}

func (mc *modbusClient) ReadCoils(offset, length uint16, unitID uint8) ([]bool, error) {
	availableRetries := 3

	for availableRetries > 0 {
		mc.mu.Lock()
		mc.client.SetUnitId(uint8(unitID))
		b, err := mc.client.ReadCoils(offset, length)
		mc.mu.Unlock()
		if err != nil {
			mc.logger.Debugf("Failed to read coils: %v", err)
			availableRetries--
			err := mc.reConnect()
			if err != nil {
				return nil, err
			}
		} else {
			return b, nil
		}
	}
	return nil, ErrRetriesExhausted
}

func (mc *modbusClient) ReadCoil(offset uint16, unitID uint8) (bool, error) {
	availableRetries := 3

	for availableRetries > 0 {
		mc.mu.Lock()
		mc.client.SetUnitId(unitID)
		b, err := mc.client.ReadCoil(offset)
		mc.mu.Unlock()
		if err != nil {
			availableRetries--
			err := mc.reConnect()
			if err != nil {
				return false, err
			}
		} else {
			return b, nil
		}
	}
	return false, ErrRetriesExhausted
}

func (mc *modbusClient) ReadDiscreteInputs(offset, length uint16, unitID uint8) ([]bool, error) {
	availableRetries := 3

	for availableRetries > 0 {
		mc.mu.Lock()
		mc.client.SetUnitId(unitID)
		b, err := mc.client.ReadDiscreteInputs(offset, length)
		mc.mu.Unlock()
		if err != nil {
			availableRetries--
			err := mc.reConnect()
			if err != nil {
				return nil, err
			}
		} else {
			return b, nil
		}
	}
	return nil, ErrRetriesExhausted
}

func (mc *modbusClient) ReadDiscreteInput(offset uint16, unitID uint8) (bool, error) {
	availableRetries := 3

	for availableRetries > 0 {
		mc.mu.Lock()
		mc.client.SetUnitId(unitID)
		b, err := mc.client.ReadDiscreteInput(offset)
		mc.mu.Unlock()
		if err != nil {
			availableRetries--
			err := mc.reConnect()
			if err != nil {
				return false, err
			}
		} else {
			return b, nil
		}
	}
	return false, ErrRetriesExhausted
}

func (mc *modbusClient) ReadHoldingRegisters(offset, length uint16, unitID uint8) ([]uint16, error) {
	availableRetries := 3

	for availableRetries > 0 {
		mc.mu.Lock()
		mc.client.SetUnitId(unitID)
		b, err := mc.client.ReadRegisters(offset, length, modbus.HOLDING_REGISTER)
		mc.mu.Unlock()
		if err != nil {
			availableRetries--
			err := mc.reConnect()
			if err != nil {
				return nil, err
			}
		} else {
			return b, nil
		}
	}
	return nil, ErrRetriesExhausted
}

func (mc *modbusClient) ReadInputRegisters(offset, length uint16, unitID uint8) ([]uint16, error) {
	availableRetries := 3

	for availableRetries > 0 {
		mc.mu.Lock()
		mc.client.SetUnitId(unitID)
		b, err := mc.client.ReadRegisters(offset, length, modbus.INPUT_REGISTER)
		mc.mu.Unlock()
		if err != nil {
			availableRetries--
			err := mc.reConnect()
			if err != nil {
				return nil, err
			}
		} else {
			return b, nil
		}
	}
	return nil, ErrRetriesExhausted
}

func (mc *modbusClient) ReadInt32(offset uint16, regType modbus.RegType, unitID uint8) (int32, error) {
	availableRetries := 3

	for availableRetries > 0 {
		mc.mu.Lock()
		mc.client.SetUnitId(unitID)
		b, err := mc.client.ReadUint32(offset, regType)
		mc.mu.Unlock()
		if err != nil {
			availableRetries--
			err := mc.reConnect()
			if err != nil {
				return 0, err
			}
		} else {
			return int32(b), nil
		}
	}
	return 0, ErrRetriesExhausted
}

func (mc *modbusClient) ReadUInt32(offset uint16, regType modbus.RegType, unitID uint8) (uint32, error) {
	availableRetries := 3

	for availableRetries > 0 {
		mc.mu.Lock()
		mc.client.SetUnitId(unitID)
		b, err := mc.client.ReadUint32(offset, regType)
		mc.mu.Unlock()
		if err != nil {
			availableRetries--
			err := mc.reConnect()
			if err != nil {
				return 0, err
			}
		} else {
			return b, nil
		}
	}
	return 0, ErrRetriesExhausted
}

func (mc *modbusClient) ReadUInt64(offset uint16, regType modbus.RegType, unitID uint8) (uint64, error) {
	availableRetries := 3

	for availableRetries > 0 {
		mc.mu.Lock()
		mc.client.SetUnitId(unitID)
		b, err := mc.client.ReadUint64(offset, regType)
		mc.mu.Unlock()
		if err != nil {
			availableRetries--
			err := mc.reConnect()
			if err != nil {
				return 0, err
			}
		} else {
			return b, nil
		}
	}
	return 0, ErrRetriesExhausted
}

func (mc *modbusClient) ReadFloat32(offset uint16, regType modbus.RegType, unitID uint8) (float32, error) {
	availableRetries := 3

	for availableRetries > 0 {
		mc.mu.Lock()
		mc.client.SetUnitId(unitID)
		b, err := mc.client.ReadFloat32(offset, regType)
		mc.mu.Unlock()
		if err != nil {
			availableRetries--
			err := mc.reConnect()
			if err != nil {
				return 0, err
			}
		} else {
			return b, nil
		}
	}
	return 0, ErrRetriesExhausted
}

func (mc *modbusClient) ReadFloat64(offset uint16, regType modbus.RegType, unitID uint8) (float64, error) {
	availableRetries := 3

	for availableRetries > 0 {
		mc.mu.Lock()
		mc.client.SetUnitId(unitID)
		b, err := mc.client.ReadFloat64(offset, regType)
		mc.mu.Unlock()
		if err != nil {
			availableRetries--
			err := mc.reConnect()
			if err != nil {
				return 0, err
			}
		} else {
			return b, nil
		}
	}
	return 0, ErrRetriesExhausted
}

func (mc *modbusClient) ReadUInt8(offset uint16, regType modbus.RegType, unitID uint8) (uint8, error) {
	availableRetries := 3

	for availableRetries > 0 {
		mc.mu.Lock()
		mc.client.SetUnitId(unitID)
		b, err := mc.client.ReadRegister(offset, regType)
		mc.mu.Unlock()
		if err != nil {
			availableRetries--
			err := mc.reConnect()
			if err != nil {
				return 0, err
			}
		} else {
			return uint8(b), nil
		}
	}
	return 0, ErrRetriesExhausted
}

func (mc *modbusClient) ReadInt16(offset uint16, regType modbus.RegType, unitID uint8) (int16, error) {
	availableRetries := 3

	for availableRetries > 0 {
		mc.mu.Lock()
		mc.client.SetUnitId(unitID)
		b, err := mc.client.ReadRegister(offset, regType)
		mc.mu.Unlock()
		if err != nil {
			availableRetries--
			err := mc.reConnect()
			if err != nil {
				return 0, err
			}
		} else {
			return int16(b), nil
		}
	}
	return 0, ErrRetriesExhausted
}

func (mc *modbusClient) ReadUInt16(offset uint16, regType modbus.RegType, unitID uint8) (uint16, error) {
	availableRetries := 3

	for availableRetries > 0 {
		mc.mu.Lock()
		mc.client.SetUnitId(unitID)
		b, err := mc.client.ReadRegister(offset, regType)
		mc.mu.Unlock()
		if err != nil {
			availableRetries--
			err := mc.reConnect()
			if err != nil {
				return 0, err
			}
		} else {
			return b, nil
		}
	}
	return 0, ErrRetriesExhausted
}

func (mc *modbusClient) ReadBytes(offset, length uint16, regType modbus.RegType, unitID uint8) ([]byte, error) {
	availableRetries := 3

	for availableRetries > 0 {
		mc.mu.Lock()
		mc.client.SetUnitId(unitID)
		b, err := mc.client.ReadBytes(offset, length, regType)
		mc.mu.Unlock()
		if err != nil {
			availableRetries--
			err := mc.reConnect()
			if err != nil {
				return nil, err
			}
		} else {
			return b, nil
		}
	}
	return nil, ErrRetriesExhausted
}

func (mc *modbusClient) ReadRawBytes(offset, length uint16, regType modbus.RegType, unitID uint8) ([]byte, error) {
	availableRetries := 3

	for availableRetries > 0 {
		mc.mu.Lock()
		mc.client.SetUnitId(unitID)
		b, err := mc.client.ReadRawBytes(offset, length, regType)
		mc.mu.Unlock()
		if err != nil {
			availableRetries--
			err := mc.reConnect()
			if err != nil {
				return nil, err
			}
		} else {
			return b, nil
		}
	}
	return nil, ErrRetriesExhausted
}

func (mc *modbusClient) WriteCoil(offset uint16, value bool, unitID uint8) error {
	availableRetries := 3

	for availableRetries > 0 {
		mc.mu.Lock()
		mc.client.SetUnitId(unitID)
		err := mc.client.WriteCoil(offset, value)
		mc.mu.Unlock()
		if err != nil {
			availableRetries--
			err := mc.reConnect()
			if err != nil {
				return err
			}
		} else {
			return nil
		}
	}
	return ErrRetriesExhausted
}

func (mc *modbusClient) WriteUInt16(offset uint16, value uint16, unitID uint8) error {
	return mc.WriteWithRetry(func() error {
		return mc.client.WriteRegister(offset, value)
	}, unitID)
}

func (mc *modbusClient) WriteUInt32(offset uint16, value uint32, unitID uint8) error {
	return mc.WriteWithRetry(func() error {
		return mc.client.WriteUint32(offset, value)
	}, unitID)
}

func (mc *modbusClient) WriteUInt64(offset uint16, value uint64, unitID uint8) error {
	return mc.WriteWithRetry(func() error {
		return mc.client.WriteUint64(offset, value)
	}, unitID)
}

func (mc *modbusClient) WriteFloat32(offset uint16, value float32, unitID uint8) error {
	return mc.WriteWithRetry(func() error {
		return mc.client.WriteFloat32(offset, value)
	}, unitID)
}

func (mc *modbusClient) WriteFloat64(offset uint16, value float64, unitID uint8) error {
	return mc.WriteWithRetry(func() error {
		return mc.client.WriteFloat64(offset, value)
	}, unitID)
}

func (mc *modbusClient) WriteWithRetry(w func() error, unitID uint8) error {
	availableRetries := 3

	for availableRetries > 0 {
		mc.mu.Lock()
		mc.client.SetUnitId(unitID)
		err := w()
		mc.mu.Unlock()
		if err != nil {
			availableRetries--
			err := mc.reConnect()
			if err != nil {
				return err
			}
		} else {
			return nil
		}
	}
	return ErrRetriesExhausted
}

func (mc *modbusClient) reConnect() error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.logger.Debugf("Re-initializing modbus client")
	err := mc.client.Open()
	if err != nil {
		mc.logger.Errorf("Failed to re-open modbus client: %#v", err)
		return err
	}
	return nil
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
