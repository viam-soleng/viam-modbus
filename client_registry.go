package viammodbus

import (
	"context"
	"fmt"
	"sync"

	"go.viam.com/rdk/logging"
)

var GlobalClientRegistry *clientRegistry = &clientRegistry{clients: map[string]Client{}}

type Client interface {
	Start() error
	Close(ctx context.Context) error
}

type clientRegistry struct {
	clients map[string]Client
	mu      sync.Mutex
}

func (cr *clientRegistry) Add(name string, c Client) error {
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

func (cr *clientRegistry) GetOrCreate(src string, logger logging.Logger) (Client, error) {
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

// TODO: Implement a default client that does nothing, so we can always return a client.
type defaultClient struct{}

func (d *defaultClient) Start() error {
	return nil
}

func (d *defaultClient) Close(ctx context.Context) error {
	return nil
}

func CreateClient(src string, logger logging.Logger) Client {
	// VIAMBOAT: https://github.com/viam-labs/viamboat/blob/83aed8a52948aa6f41e0977cbc9f07921a6a801b/reader.go#L269
	return &defaultClient{}
}
