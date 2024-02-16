package modbus_tcp_board

import "github.com/simonvetter/modbus"

func reinitializeModbusClient(r *ModbusTcpBoard) error {
	r.logger.Warnf("Re-initializing modbus client")
	return initializeModbusClient(r)
}

func initializeModbusClient(r *ModbusTcpBoard) error {
	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     r.uri,
		Timeout: r.timeout,
	})
	if err != nil {
		r.logger.Errorf("Failed to create modbus client: %#v", err)
		return err
	}
	err = client.Open()
	if err != nil {
		r.logger.Errorf("Failed to open modbus client: %#v", err)
		return err
	}
	r.client = client
	return nil
}

func executeWithRetry(r *ModbusTcpBoard, f func() (int, error)) (int, error) {
	availableRetries := 3
retry:
	i, err := f()
	if err != nil && availableRetries > 0 {
		availableRetries--
		err := reinitializeModbusClient(r)
		if err != nil {
			return 0, err
		}
		goto retry
	}
	return i, err
}
