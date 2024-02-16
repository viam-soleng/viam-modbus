package modbus_tcp

import (
	"testing"
	"time"

	"github.com/simonvetter/modbus"
	"github.com/stretchr/testify/assert"
)

func TestGetBytes(t *testing.T) {
	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     "tcp://10.86.5.41:502",
		Timeout: 10 * time.Second,
	})
	assert.NoError(t, err)
	err = client.Open()
	assert.NoError(t, err)
	defer client.Close()
	length := 1
	for offset := 0; offset < 100; offset += length {
		b, err := client.ReadCoils(uint16(offset), uint16(length))
		assert.NoError(t, err)
		t.Logf("ReadRegisters: %v", b)
	}
}
