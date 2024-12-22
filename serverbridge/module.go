// Package serverbridge allows for multiple different modbus clients to be bridged together over different protocols.
package serverbridge

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/goburrow/serial"
	"github.com/rinzlerlabs/gomodbus/server"
	"github.com/rinzlerlabs/gomodbus/server/network/tcp"
	"github.com/rinzlerlabs/gomodbus/server/serial/ascii"
	"github.com/rinzlerlabs/gomodbus/server/serial/rtu"
	"go.viam.com/rdk/components/board"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"

	"viam-modbus/common"
)

type resourceType = sensor.Sensor
type config = *modbusBridgeConfig

var API = board.API
var Model = resource.NewModel("viam-soleng", "modbus", "server-bridge")
var PrettyName = "Modbus Server Bridge"

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
		Named:      conf.ResourceName().AsNamed(),
		logger:     logger,
		cancelFunc: cancelFunc,
		ctx:        c,
		handler:    server.NewDefaultHandler(logger.AsZap().Desugar(), 65535, 65535, 65535, 65535),
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
	logger     logging.Logger
	cancelFunc context.CancelFunc
	ctx        context.Context
	handler    server.RequestHandler
	servers    map[string]server.ModbusServer
}

func (b *modbusBridge) Reconfigure(ctx context.Context, deps resource.Dependencies, rawConf resource.Config) error {
	b.logger.Infof("Reconfiguring %v Component %v", PrettyName, common.Version)

	conf, err := resource.NativeConfig[config](rawConf)
	if err != nil {
		return err
	}

	servers := make(map[string]server.ModbusServer)
	errs := make([]error, 0)
	for _, endpoint := range conf.Endpoints {
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
				server, err = rtu.NewModbusServer(b.logger.AsZap().Desugar(), port, endpoint.SerialConfig.ServerAddress)
			} else {
				server, err = ascii.NewModbusServer(b.logger.AsZap().Desugar(), port, endpoint.SerialConfig.ServerAddress)
			}
			if err != nil {
				errs = append(errs, err)
			} else {
				servers[endpoint.Name] = server
			}
		} else if endpoint.TCPConfig != nil {
			server, err := tcp.NewModbusServer(b.logger.Desugar(), endpoint.Endpoint)
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
	return startServers(servers)
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
		stopServers(servers)
	}
	return errors.Join(errs...)
}

func stopServers(servers map[string]server.ModbusServer) error {
	var errs error
	for _, s := range servers {
		err := s.Stop()
		if err != nil {
			errors.Join(errs, err)
		}
	}
	return errs
}

func (b *modbusBridge) Close(ctx context.Context) error {
	return stopServers(b.servers)
}

func (b *modbusBridge) Readings(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	ret := make(map[string]interface{})
	ret["ActiveServers"] = len(b.servers)
	for serverName, server := range b.servers {
		addStats(ret, serverName, server)
	}

	return nil, nil
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
	m[fmt.Sprintf("%s.%s", serverName, statName)] = stat
}
