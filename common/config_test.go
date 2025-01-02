package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *ModbusConfig
		wantErr error
	}{
		{
			name:    "empty config",
			cfg:     &ModbusConfig{},
			wantErr: ErrEndpointMustBeSet,
		},
		{
			name:    "invalid log level",
			cfg:     &ModbusConfig{Endpoint: "tcp://localhost:502", LogLevel: "invalid"},
			wantErr: ErrInvalidLogLevel,
		},
		{
			name: "debug log level",
			cfg:  &ModbusConfig{Endpoint: "tcp://localhost:502", LogLevel: "debug"},
		},
		{
			name: "info log level",
			cfg:  &ModbusConfig{Endpoint: "tcp://localhost:502", LogLevel: "info"},
		},
		{
			name: "warn log level",
			cfg:  &ModbusConfig{Endpoint: "tcp://localhost:502", LogLevel: "warn"},
		},
		{
			name: "error log level",
			cfg:  &ModbusConfig{Endpoint: "tcp://localhost:502", LogLevel: "error"},
		},
		{
			name:    "invalid scheme",
			cfg:     &ModbusConfig{Endpoint: "foo://localhost:502", LogLevel: "error"},
			wantErr: ErrInvalidScheme,
		},
		{
			name:    "ascii invalid device",
			cfg:     &ModbusConfig{Endpoint: "ascii:///dev/ttyUSB12", LogLevel: "error"},
			wantErr: ErrNoSuchDevice,
		},
		{
			name:    "ascii invalid speed",
			cfg:     &ModbusConfig{Endpoint: "ascii:///dev/tty", LogLevel: "error", Speed: 0},
			wantErr: ErrSpeedMustBeNonZero,
		},
		{
			name:    "ascii invalid parity 1",
			cfg:     &ModbusConfig{Endpoint: "ascii:///dev/tty", LogLevel: "error", Speed: 19200},
			wantErr: ErrInvalidParity,
		},
		{
			name:    "ascii invalid parity 2",
			cfg:     &ModbusConfig{Endpoint: "ascii:///dev/tty", LogLevel: "error", Speed: 19200, Parity: "A"},
			wantErr: ErrInvalidParity,
		},
		{
			name:    "ascii invalid data bits",
			cfg:     &ModbusConfig{Endpoint: "ascii:///dev/tty", LogLevel: "error", Speed: 19200, Parity: "O", DataBits: 8},
			wantErr: ErrInvalidASCIIDataBits,
		},
		{
			name:    "ascii invalid stop bits",
			cfg:     &ModbusConfig{Endpoint: "ascii:///dev/tty", LogLevel: "error", Speed: 19200, Parity: "O", DataBits: 7},
			wantErr: ErrInvalidASCIIStopBits,
		},
		{
			name: "ascii valid config",
			cfg:  &ModbusConfig{Endpoint: "ascii:///dev/tty", LogLevel: "error", Speed: 19200, Parity: "O", DataBits: 7, StopBits: 1},
		},
		{
			name:    "rtu invalid speed",
			cfg:     &ModbusConfig{Endpoint: "rtu:///dev/tty", LogLevel: "error"},
			wantErr: ErrSpeedMustBeNonZero,
		},
		{
			name:    "rtu invalid parity 1",
			cfg:     &ModbusConfig{Endpoint: "rtu:///dev/tty", LogLevel: "error", Speed: 19200},
			wantErr: ErrInvalidParity,
		},
		{
			name:    "rtu invalid parity 2",
			cfg:     &ModbusConfig{Endpoint: "rtu:///dev/tty", LogLevel: "error", Speed: 19200, Parity: "A"},
			wantErr: ErrInvalidParity,
		},
		{
			name:    "rtu invalid data bits",
			cfg:     &ModbusConfig{Endpoint: "rtu:///dev/tty", LogLevel: "error", Speed: 19200, Parity: "O"},
			wantErr: ErrInvalidRTUDataBits,
		},
		{
			name:    "rtu invalid stop bits",
			cfg:     &ModbusConfig{Endpoint: "rtu:///dev/tty", LogLevel: "error", Speed: 19200, Parity: "O", DataBits: 8},
			wantErr: ErrInvalidRTUStopBits,
		},
		{
			name: "rtu valid config",
			cfg:  &ModbusConfig{Endpoint: "rtu:///dev/tty", LogLevel: "error", Speed: 19200, Parity: "O", DataBits: 8, StopBits: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.cfg.Validate(); err != tt.wantErr {
				assert.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}
