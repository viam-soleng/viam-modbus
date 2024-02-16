package modbus_tcp_board

import (
	"context"
	"errors"
	"sync"

	"github.com/simonvetter/modbus"
)

type ModbusAnlogPin struct {
	mu       *sync.RWMutex
	offset   uint16
	dataType string
	pinType  modbus.RegType
	board    *ModbusTcpBoard
}

// Close implements board.AnalogReader.
func (*ModbusAnlogPin) Close(ctx context.Context) error {
	// Do we need to do anything here?
	return nil
}

// Read implements board.AnalogReader.
func (r *ModbusAnlogPin) Read(ctx context.Context, extra map[string]interface{}) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	switch r.dataType {
	case "uint32":
		return executeWithRetry(r.board, func() (int, error) {
			val, err := r.board.client.ReadUint32(r.offset, r.pinType)
			return int(val), err
		})
	case "uint64":
		return executeWithRetry(r.board, func() (int, error) {
			val, err := r.board.client.ReadUint64(r.offset, r.pinType)
			return int(val), err
		})
	case "float32":
		return executeWithRetry(r.board, func() (int, error) {
			val, err := r.board.client.ReadFloat32(r.offset, r.pinType)
			return int(val), err
		})
	case "float64":
		return executeWithRetry(r.board, func() (int, error) {
			val, err := r.board.client.ReadFloat64(r.offset, r.pinType)
			return int(val), err
		})
	default:
		return 0, errors.New("invalid data type")
	}
}

// Why do we need a board here? because we need to reinitialize the modbus client if there is an error. this can probably be done better, but here we are
func NewModbusAnalogPin(board *ModbusTcpBoard, conf ModbusAnalogPinConfig) (*ModbusAnlogPin, error) {
	var t modbus.RegType
	if conf.PinType == "input" {
		t = modbus.INPUT_REGISTER
	} else if conf.PinType == "holding" {
		t = modbus.HOLDING_REGISTER
	} else {
		return nil, errors.New("invalid pin type")
	}
	return &ModbusAnlogPin{
		mu:       &board.mu,
		offset:   uint16(conf.Offset),
		dataType: conf.DataType,
		pinType:  t,
		board:    board,
	}, nil
}
