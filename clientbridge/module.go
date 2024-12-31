// Package clientbridge allows for multiple different modbus servers to be bridged together over different protocols.
package clientbridge

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/rinzlerlabs/gomodbus/client"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"go.viam.com/utils"

	"viam-modbus/common"
)

type resourceType = sensor.Sensor
type config = *modbusBridgeConfig

var (
	API        = sensor.API
	Model      = resource.NewModel("viam-soleng", "modbus", "client-bridge")
	PrettyName = "Modbus Client Bridge"
)

func init() {
	resource.RegisterComponent(
		API,
		Model,
		resource.Registration[resourceType, config]{
			Constructor: NewModbusBridge,
		},
	)
}

func NewModbusBridge(ctx context.Context, deps resource.Dependencies, conf resource.Config, logger logging.Logger) (resourceType, error) {
	logger.Infof("Starting Modbus Board Component %v", common.Version)
	b := modbusBridge{
		Named:  conf.ResourceName().AsNamed(),
		logger: logger,
	}

	if err := b.Reconfigure(ctx, deps, conf); err != nil {
		logger.Errorf("Failed to start Modbus Board Component %v", err)
		return nil, err
	}
	logger.Info("Modbus Board Component started successfully")
	return &b, nil
}

type modbusBridge struct {
	resource.Named
	logger  logging.Logger
	mu      sync.Mutex
	clients map[string]*addressableClient
	workers *utils.StoppableWorkers
}

func (b *modbusBridge) Reconfigure(ctx context.Context, deps resource.Dependencies, rawConf resource.Config) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	var err error
	b.logger.Infof("Reconfiguring %v Component %v", PrettyName, common.Version)
	if b.workers != nil {
		b.workers.Stop()
	}

	conf, err := resource.NativeConfig[config](rawConf)
	if err != nil {
		return err
	}

	clients := make(map[string]*addressableClient)
	errs := make([]error, 0)
	for _, endpoint := range conf.Endpoints {
		client, err := common.NewModbusClientFromConfig(b.logger, endpoint)
		if err != nil {
			errs = append(errs, err)
		} else {
			clients[endpoint.Name] = &addressableClient{client, endpoint.ServerAddress}
		}
	}

	b.clients = clients
	if len(errs) > 0 {
		err := errors.Join(errs...)
		b.logger.Errorf("Failed to create clients: %v", err)
		e := b.close()
		b.logger.Errorf("Failed to close clients: %v", e)
		return err
	}

	sleepTime := time.Duration(conf.UpdateTimeMs) * time.Millisecond
	workers := make([]func(context.Context), 0)
	for _, block := range conf.Blocks {
		if b.logger == nil {
			b.close()
			return errors.New("logger is nil")
		}
		if block.Src == "" {
			b.close()
			return errors.New("source is empty")
		}
		if block.Dst == "" {
			b.close()
			return errors.New("destination is empty")
		}
		if b.clients[block.Src] == nil {
			b.close()
			return errors.New("source client is nil")
		}
		if b.clients[block.Dst] == nil {
			b.close()
			return errors.New("destination client is nil")
		}
		w := newModbusClientBridge(b.logger, b.clients[block.Src], b.clients[block.Dst], block, sleepTime)
		workers = append(workers, w.Start)
	}

	b.workers = utils.NewBackgroundStoppableWorkers(workers...)

	return nil
}

func (b *modbusBridge) close() error {
	errs := make([]error, 0)
	for _, client := range b.clients {
		errs = append(errs, client.Close())
	}
	return errors.Join(errs...)
}

func (b *modbusBridge) Close(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.logger.Infof("Closing %s Component", PrettyName)
	b.workers.Stop()
	if err := b.close(); err != nil {
		b.logger.Errorf("Failed to close %s Component: %v", PrettyName, err)
	}
	return nil
}

func (b *modbusBridge) Readings(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return nil, nil
}

type addressableClient struct {
	client.ModbusClient
	address uint16
}

type modbusClientBridge struct {
	logger        logging.Logger
	sourceClient  client.ModbusClient
	sourceAddress uint16
	destClient    client.ModbusClient
	destAddress   uint16
	block         modbusBlock
	updateTime    time.Duration
}

func newModbusClientBridge(logger logging.Logger, sourceClient, destClient *addressableClient, block modbusBlock, updateTime time.Duration) *modbusClientBridge {
	return &modbusClientBridge{
		logger:        logger,
		sourceClient:  sourceClient,
		sourceAddress: sourceClient.address,
		destClient:    destClient,
		destAddress:   destClient.address,
		block:         block,
		updateTime:    updateTime,
	}
}

func (b *modbusClientBridge) Start(ctx context.Context) {
	if b.logger == nil {
		return
	}
	if ctx == nil {
		b.logger.Warn("Context is nil")
		return
	}
	b.logger.Info("Starting bridge Worker Src: %s Dst: %s", b.block.Src, b.block.Dst)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(b.updateTime):
			b.iterate()
		}
	}
}

func (b *modbusClientBridge) iterate() {
	b.logger.Info("Starting iteration")
	block := b.block
	switch block.SrcRegister {
	case "coils":
		reg, err := b.sourceClient.ReadCoils(b.sourceAddress, uint16(block.SrcOffset), uint16(block.Length))
		if err != nil {
			b.logger.Errorf("Failed to read coils: %v", err)
			return
		}
		err = b.destClient.WriteMultipleCoils(b.destAddress, uint16(block.DstOffset), reg)
		if err != nil {
			b.logger.Errorf("Failed to write coils: %v", err)
			return
		}
	case "discrete_inputs":
		reg, err := b.sourceClient.ReadDiscreteInputs(b.sourceAddress, uint16(block.SrcOffset), uint16(block.Length))
		if err != nil {
			b.logger.Errorf("Failed to read discrete inputs: %v", err)
			return
		}
		err = b.destClient.WriteMultipleCoils(b.destAddress, uint16(block.DstOffset), reg)
		if err != nil {
			b.logger.Errorf("Failed to write discrete inputs: %v", err)
			return
		}
	case "holding_registers":
		reg, err := b.sourceClient.ReadHoldingRegisters(b.sourceAddress, uint16(block.SrcOffset), uint16(block.Length))
		if err != nil {
			b.logger.Errorf("Failed to read holding registers: %v", err)
			return
		}
		err = b.destClient.WriteMultipleRegisters(b.destAddress, uint16(block.DstOffset), reg)
		if err != nil {
			b.logger.Errorf("Failed to write holding registers: %v", err)
			return
		}
	case "input_registers":
		reg, err := b.sourceClient.ReadInputRegisters(b.sourceAddress, uint16(block.SrcOffset), uint16(block.Length))
		if err != nil {
			b.logger.Errorf("Failed to read input registers: %v", err)
			return
		}
		err = b.destClient.WriteMultipleRegisters(b.destAddress, uint16(block.DstOffset), reg)
		if err != nil {
			b.logger.Errorf("Failed to write input registers: %v", err)
			return
		}
	}
}
