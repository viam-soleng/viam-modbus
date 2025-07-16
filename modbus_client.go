package viammodbus

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/simonvetter/modbus"
	"github.com/viam-soleng/viam-modbus/common"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/services/generic"
)

var NamespaceFamily = resource.NewModelFamily("viam-soleng", "modbus")
var ModbusClientModel = NamespaceFamily.WithModel("clients")

func init() {
	resource.RegisterComponent(
		generic.API,
		ModbusClientModel,
		resource.Registration[generic.Service, *modbusClientConfig]{
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
	TLSClientCert string `json:"tls_client_cert"`
	TLSRootCAs    string `json:"tls_root_cas"`
	Endianness    string `json:"endianness"`
	WordOrder     string `json:"word_order"`
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
	return []string{}, nil, nil
}

type modbusClient struct {
	//TODO: Check if always rebuild works as original module did use reconfig
	resource.AlwaysRebuild

	name   resource.Name
	mu     sync.RWMutex
	logger logging.Logger

	endianness modbus.Endianness
	wordOrder  modbus.WordOrder
}

func newModbusClient(ctx context.Context, deps resource.Dependencies, config resource.Config, logger logging.Logger) (generic.Service, error) {
	newConf, err := resource.NativeConfig[*modbusClientConfig](config)
	if err != nil {
		return nil, err
	}

	endianness, err := common.GetEndianness(newConf.Endianness)
	if err != nil {
		return nil, err
	}

	wordOrder, err := common.GetWordOrder(newConf.WordOrder)
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
		// TODO: To be implemented
		//TLSClientCert: tlsClientCert,
		//TLSRootCAs:    tlsRootCAs,
	}

	return NewModbusClient(ctx, config.ResourceName(), endianness, wordOrder, clientConfig, logger)
}

func NewModbusClient(ctx context.Context, name resource.Name, endianness modbus.Endianness, wordOrder modbus.WordOrder, clientConfig modbus.ClientConfiguration, logger logging.Logger) (generic.Service, error) {
	client := &modbusClient{
		logger:     logger,
		endianness: endianness,
		wordOrder:  wordOrder,
	}

	//TODO: Put into seperate function e.g. connect()
	/*
		err := client.initializeModbusClient()
		if err != nil {
			return nil, err
		}
	*/

	err := GlobalClientRegistry.Add(name.ShortName(), client)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// Start implements the Client interface.
func (gs *modbusClient) Start() error {
	// Add initialization logic here if needed
	return nil
}

func (gs *modbusClient) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

func (gs *modbusClient) Close(ctx context.Context) error {
	return nil
}

func (gs *modbusClient) Name() resource.Name {
	return gs.name
}
