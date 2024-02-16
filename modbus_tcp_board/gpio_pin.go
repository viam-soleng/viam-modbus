package modbus_tcp_board

import (
	"context"
	"errors"
	"sync"
)

var OUTPUT_PIN = "output"
var INPUT_PIN = "input"

type ModbusGpioPin struct {
	mu     *sync.RWMutex
	offset uint16
	board  *ModbusTcpBoard
}

// Get implements board.GPIOPin.
func (r *ModbusGpioPin) Get(ctx context.Context, extra map[string]interface{}) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	i, e := executeWithRetry(r.board, func() (int, error) {
		b, e := r.board.client.ReadCoil(r.offset)
		if b {
			return 1, e
		} else {
			return 0, e
		}
	})

	// This is kind of dumb, but it gets the job done without having to duplicate the executeWithRetry function for bools
	if i == 1 {
		return true, e
	} else {
		return false, e
	}
}

// PWM implements board.GPIOPin.
func (*ModbusGpioPin) PWM(ctx context.Context, extra map[string]interface{}) (float64, error) {
	return 0, errors.ErrUnsupported
}

// PWMFreq implements board.GPIOPin.
func (*ModbusGpioPin) PWMFreq(ctx context.Context, extra map[string]interface{}) (uint, error) {
	return 0, errors.ErrUnsupported
}

// Set implements board.GPIOPin.
func (r *ModbusGpioPin) Set(ctx context.Context, high bool, extra map[string]interface{}) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, e := executeWithRetry(r.board, func() (int, error) {
		err := r.board.client.WriteCoil(r.offset, high)
		return 0, err
	})
	return e
}

// SetPWM implements board.GPIOPin.
func (*ModbusGpioPin) SetPWM(ctx context.Context, dutyCyclePct float64, extra map[string]interface{}) error {
	return errors.ErrUnsupported
}

// SetPWMFreq implements board.GPIOPin.
func (*ModbusGpioPin) SetPWMFreq(ctx context.Context, freqHz uint, extra map[string]interface{}) error {
	return errors.ErrUnsupported
}

func NewModbusGpioPin(board *ModbusTcpBoard, conf ModbusGpioPinConfig) *ModbusGpioPin {
	return &ModbusGpioPin{
		mu:     &board.mu,
		offset: uint16(conf.Offset),
		board:  board,
	}
}
