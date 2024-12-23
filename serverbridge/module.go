// Package serverbridge allows for multiple different modbus clients to be bridged together over different protocols.
package serverbridge

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/goburrow/serial"
	"github.com/rinzlerlabs/gomodbus/server"
	"github.com/rinzlerlabs/gomodbus/server/network/tcp"
	"github.com/rinzlerlabs/gomodbus/server/serial/ascii"
	"github.com/rinzlerlabs/gomodbus/server/serial/rtu"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"

	"viam-modbus/common"
)

type resourceType = sensor.Sensor
type config = *modbusBridgeConfig

var (
	ErrNoDataDir = errors.New("no data directory")
	API          = sensor.API
	Model        = resource.NewModel("viam-soleng", "modbus", "server-bridge")
	PrettyName   = "Modbus Server Bridge"
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
	logger.Infof("Starting %v Component %v", PrettyName, common.Version)
	c, cancelFunc := context.WithCancel(context.Background())
	b := modbusBridge{
		Named:       conf.ResourceName().AsNamed(),
		logger:      logger,
		cancelFunc:  cancelFunc,
		ctx:         c,
		handler:     server.NewDefaultHandler(logger.AsZap().Desugar(), 65535, 65535, 65535, 65535),
		persistData: false,
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
	logger      logging.Logger
	cancelFunc  context.CancelFunc
	ctx         context.Context
	handler     server.RequestHandler
	servers     map[string]server.ModbusServer
	persistData bool
	mu          sync.Mutex
}

func (b *modbusBridge) Reconfigure(ctx context.Context, deps resource.Dependencies, rawConf resource.Config) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.logger.Infof("Reconfiguring %v Component %v", PrettyName, common.Version)

	conf, err := resource.NativeConfig[config](rawConf)
	if err != nil {
		return err
	}

	b.persistData = conf.PersistData
	if b.persistData {
		b.logger.Infof("Loading persisted data")
		dataDir := os.Getenv("VIAM_MODULE_DATA")
		if _, err := os.Stat(dataDir); os.IsNotExist(err) || dataDir == "" {
			b.logger.Warnf("Data directory does not exist, cannot persist current data: %v", dataDir)
			return ErrNoDataDir
		}
		err := loadHandlerData(dataDir, b.handler)
		if err != nil {
			b.logger.Warnf("Failed to load handler data: %v", err)
			return err
		} else {
			b.logger.Infof("Data loaded successfully")
		}
	}
	servers := make(map[string]server.ModbusServer)
	errs := make([]error, 0)
	for _, endpoint := range conf.Servers {
		if endpoint.SerialConfig != nil {
			serialConfig := &serial.Config{
				Address:  endpoint.Endpoint,
				BaudRate: int(endpoint.SerialConfig.Speed),
				DataBits: int(endpoint.SerialConfig.DataBits),
				StopBits: int(endpoint.SerialConfig.StopBits),
				Parity:   endpoint.SerialConfig.Parity,
			}
			port, err := serial.Open(serialConfig)
			if err != nil {
				errs = append(errs, err)
				break
			}
			var server server.ModbusServer
			if endpoint.SerialConfig.RTU {
				server, err = rtu.NewModbusServerWithHandler(b.logger.Desugar(), port, endpoint.SerialConfig.ServerAddress, b.handler)
			} else {
				server, err = ascii.NewModbusServerWithHandler(b.logger.Desugar(), port, endpoint.SerialConfig.ServerAddress, b.handler)
			}
			if err != nil {
				errs = append(errs, err)
			} else {
				servers[endpoint.Name] = server
			}
		} else if endpoint.TCPConfig != nil {
			server, err := tcp.NewModbusServerWithHandler(b.logger.Desugar(), endpoint.Endpoint, b.handler)
			if err != nil {
				errs = append(errs, err)
			} else {
				servers[endpoint.Name] = server
			}
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	b.servers = servers
	return startServers(b.servers)
}

func startServers(servers map[string]server.ModbusServer) error {
	errs := make([]error, 0)
	for _, s := range servers {
		err := s.Start()
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		closeServers(servers)
	}
	return errors.Join(errs...)
}

func closeServers(servers map[string]server.ModbusServer) error {
	var errs error
	for _, s := range servers {
		err := s.Close()
		if err != nil {
			errors.Join(errs, err)
		}
	}
	return errs
}

func (b *modbusBridge) Close(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	err := closeServers(b.servers)
	if err != nil {
		b.logger.Warnf("Failed to stop servers: %v", err)
	}
	if b.persistData {
		b.logger.Infof("Persisting data")
		dataDir := os.Getenv("VIAM_MODULE_DATA")
		if _, err := os.Stat(dataDir); os.IsNotExist(err) || dataDir == "" {
			b.logger.Warnf("Data directory does not exist, cannot persist current data: %v", dataDir)
			return ErrNoDataDir
		}
		err = saveHandlerData(dataDir, b.handler)
		if err != nil {
			b.logger.Warnf("Failed to save handler data: %v", err)
		} else {
			b.logger.Infof("Data persisted successfully")
		}
	}
	return nil
}

func saveHandlerData(dataDir string, handler server.RequestHandler) error {
	if h, ok := handler.(server.PersistableRequestHandler); ok {
		return h.Save(dataDir)
	}
	return errors.New("handler does not support saving")
}

func loadHandlerData(dataDir string, handler server.RequestHandler) error {
	if h, ok := handler.(server.PersistableRequestHandler); ok {
		return h.Load(dataDir)
	}
	return errors.New("handler does not support loading")
}

func (b *modbusBridge) Readings(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	ret := make(map[string]interface{})
	ret["ActiveServers"] = len(b.servers)
	for serverName, server := range b.servers {
		addStats(ret, serverName, server)
	}

	return ret, nil
}

func addStats(m map[string]interface{}, serverName string, server server.ModbusServer) {
	statsMap := server.Stats().AsMap()
	for k, v := range statsMap {
		addStatKey(m, serverName, k, v)
	}

}

func addStatKey(m map[string]interface{}, serverName, statName string, stat interface{}) {
	switch v := stat.(type) {
	case int, int32, int64, float32, float64, string, bool:
		m[fmt.Sprintf("%s.%s", serverName, statName)] = v
	case []error:
		errStrings := make([]string, len(v))
		for i, err := range v {
			errStrings[i] = err.Error()
		}
		m[fmt.Sprintf("%s.%s", serverName, statName)] = strings.Join(errStrings, "\r\n")
	default:
		m[fmt.Sprintf("%s.%s", serverName, statName)] = fmt.Sprintf("%v", v)
	}
}
