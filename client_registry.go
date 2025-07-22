package viammodbus

import (
	"fmt"
	"sync"
)

var GlobalClientRegistry *clientRegistry = &clientRegistry{clients: map[string]*modbusClient{}}

type clientRegistry struct {
	clients map[string]*modbusClient
	mu      sync.Mutex
}

func (cr *clientRegistry) Add(name string, client *modbusClient) error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	existingClient, got := cr.clients[name]
	if got {
		existingClient.client.Close()
		delete(cr.clients, name)
	}
	cr.clients[name] = client
	return nil
}

func (cr *clientRegistry) Remove(name string) error {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	if client, got := cr.clients[name]; got {
		client.client.Close()
		delete(cr.clients, name)
	}
	return fmt.Errorf("no client with name [%s] found", name)
}

func (cr *clientRegistry) Get(name string) (*modbusClient, error) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	c, got := cr.clients[name]
	if !got {
		return nil, fmt.Errorf("no client with name [%s] found", name)
	}
	return c, nil
}
