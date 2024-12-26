// Package serverbridge allows for multiple different modbus clients to be bridged together over different protocols.
package serverbridge

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/goburrow/serial"
	"github.com/rinzlerlabs/gomodbus/server"
	"github.com/rinzlerlabs/gomodbus/server/network"
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
	logger.Infof("Starting %s Component %v", PrettyName, common.Version)
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
		logger.Errorf("Failed to start %s Component %v", PrettyName, err)
		return nil, err
	}
	logger.Infof("%s Component started successfully", PrettyName)
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
		if endpoint.IsSerial() {
			u, e := url.Parse(endpoint.Endpoint)
			if e != nil {
				return e
			}
			b.logger.Infof("Creating serial port %v", u.Path)
			serialConfig := &serial.Config{
				Address:  u.Path,
				BaudRate: int(endpoint.Speed),
				DataBits: int(endpoint.DataBits),
				StopBits: int(endpoint.StopBits),
				Parity:   endpoint.Parity,
			}
			var server server.ModbusServer
			if endpoint.IsRTU() {
				server, err = rtu.NewModbusServerWithHandler(b.logger.Desugar(), serialConfig, endpoint.ServerAddress, b.handler)
			} else {
				server, err = ascii.NewModbusServerWithHandler(b.logger.Desugar(), serialConfig, endpoint.ServerAddress, b.handler)
			}
			if err != nil {
				errs = append(errs, err)
			} else {
				servers[endpoint.Name] = server
			}
		} else if endpoint.IsNetwork() {
			server, err := network.NewModbusServerWithHandler(b.logger.Desugar(), endpoint.Endpoint, b.handler)
			if err != nil {
				errs = append(errs, err)
			} else {
				servers[endpoint.Name] = server
			}
		}
	}
	b.servers = servers
	if len(errs) > 0 {
		// If we had errors, we have to close servers that successfully opened because they may have acquired resources during construction.
		b.closeServers()
		return errors.Join(errs...)
	}
	return b.startServers()
}

func (b *modbusBridge) startServers() error {
	b.logger.Infof("Starting %v servers", len(b.servers))
	errs := make([]error, 0)
	for _, s := range b.servers {
		err := s.Start()
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		b.closeServers()
	}
	b.logger.Infof("Started %v servers with %v errors", len(b.servers), len(errs))
	return errors.Join(errs...)
}

func (b *modbusBridge) closeServers() error {
	b.logger.Infof("Stopping %v servers", len(b.servers))
	errs := make([]error, 0)
	for _, s := range b.servers {
		err := s.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}
	b.logger.Infof("Stopped %v servers with %v errors", len(b.servers), len(errs))
	return errors.Join(errs...)
}

func (b *modbusBridge) Close(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	err := b.closeServers()
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
