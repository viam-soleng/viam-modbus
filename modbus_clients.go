package viammodbus

import (
	"context"
	"fmt"
	"sync"

	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/services/generic"
)

var GlobalClientRegistry *clientRegistry = &clientRegistry{clients: map[string]*Client{}}

var NamespaceFamily = resource.NewModelFamily("viam-soleng", "modbus")
var ModbusClientsModel = NamespaceFamily.WithModel("clients")

func init() {
	resource.RegisterComponent(
		generic.API,
		ModbusClientsModel,
		resource.Registration[generic.Service, resource.NoNativeConfig]{
			Constructor: newModbusClientsModel,
		})
}

func newModbusClientsModel(ctx context.Context, deps resource.Dependencies, config resource.Config, logger logging.Logger) (generic.Service, error) {
	// https://github.com/viam-labs/viamboat/blob/83aed8a52948aa6f41e0977cbc9f07921a6a801b/all_pgn_sensor.go#L24C1-L32C2
	_, err := GlobalClientRegistry.GetOrCreate(config.Attributes.String("client"), logger)
	if err != nil {
		return nil, err
	}

	g := &connectionsService{
		name:   config.ResourceName(),
		logger: logger,
	}

	return g, nil
}

type connectionsService struct {
	resource.AlwaysRebuild

	name   resource.Name
	logger logging.Logger
}

type Client struct {
	logger logging.Logger
	/* Viamboat
	AddCallback(pgn int, cb ReaderCallback) // pgn or -1 for all
	Send(msg CANMessage) error
	Start() error
	Close(ctx context.Context) error
	*/
}

type clientRegistry struct {
	clients map[string]*Client
	mu      sync.Mutex
}

func (cr *clientRegistry) Add(name string, c *Client) error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	_, got := cr.clients[name]
	if got {
		return fmt.Errorf("cannot add duplicate readers with name [%s]", name)
	}

	cr.clients[name] = c
	return nil
}

func (cr *clientRegistry) Remove(name string) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	delete(cr.clients, name)
}

func (cr *clientRegistry) GetOrCreate(src string, logger logging.Logger) (*Client, error) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	c, got := cr.clients[src]
	if !got {
		if src == "" {
			return nil, fmt.Errorf("tried to create client with blank name")
		}
		c = CreateClient(src, logger)
		cr.clients[src] = c
		err := c.Start()
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (c *Client) Start() error {
	return nil
}

func CreateClient(src string, logger logging.Logger) *Client {
	// VIAMBOAT: https://github.com/viam-labs/viamboat/blob/83aed8a52948aa6f41e0977cbc9f07921a6a801b/reader.go#L269
	return &Client{
		logger: logger,
	}
}

func (gs *connectionsService) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

func (gs *connectionsService) Close(ctx context.Context) error {
	return nil
}

func (gs *connectionsService) Name() resource.Name {
	return gs.name
}
