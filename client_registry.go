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

	client, got := cr.clients[name]
	if got {
		//TODO: Make sure the mobus client is closed before replacing it
	}
	cr.clients[name] = client
	return nil
}

func (cr *clientRegistry) Remove(name string) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	if _, got := cr.clients[name]; got {
		// TODO: Close the client if it is not already closed
	}
	delete(cr.clients, name)
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
