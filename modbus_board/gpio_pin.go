package modbus_board

import (
	"context"
	"errors"
	"sync"

	"viam-modbus/common"
)

type ModbusGpioPin struct {
	mu      *sync.RWMutex
	offset  uint16
	board   *ModbusBoard
	pinType common.PinType
}

// Get implements board.GPIOPin.
func (r *ModbusGpioPin) Get(ctx context.Context, extra map[string]interface{}) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// TODO: Add Unit ID functionality
	const unitID = int8(-1) // -1 means we don't set unit ID

	if r.pinType == common.INPUT_PIN {
		return r.board.client.ReadDiscreteInput(r.offset, unitID)
	} else {
		return r.board.client.ReadCoil(r.offset, unitID)
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

	// TODO: Add Unit ID functionality
	const unitID = int8(-1) // -1 means we don't set unit ID

	if r.pinType == common.INPUT_PIN {
		return common.ErrSetInputPin
	}
	return r.board.client.WriteCoil(r.offset, high, unitID)
}

// SetPWM implements board.GPIOPin.
func (*ModbusGpioPin) SetPWM(ctx context.Context, dutyCyclePct float64, extra map[string]interface{}) error {
	return errors.ErrUnsupported
}

// SetPWMFreq implements board.GPIOPin.
func (*ModbusGpioPin) SetPWMFreq(ctx context.Context, freqHz uint, extra map[string]interface{}) error {
	return errors.ErrUnsupported
}

func NewModbusGpioPin(board *ModbusBoard, offset uint16, pinType common.PinType) *ModbusGpioPin {
	return &ModbusGpioPin{
		mu:      &board.mu,
		offset:  offset,
		board:   board,
		pinType: pinType,
	}
}
