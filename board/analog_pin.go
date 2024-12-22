package board

import (
	"context"
	"errors"
	"sync"

	"go.viam.com/rdk/components/board"

	"viam-modbus/common"
)

type ModbusAnalogPin struct {
	mu       *sync.RWMutex
	offset   uint16
	dataType string
	pinType  common.RegisterType
	board    *ModbusBoard
}

// Close implements board.AnalogReader.
func (*ModbusAnalogPin) Close(ctx context.Context) error {
	// Do we need to do anything here?
	return nil
}

// Read implements board.AnalogReader.
func (r *ModbusAnalogPin) Read(ctx context.Context, extra map[string]interface{}) (board.AnalogValue, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	switch r.dataType {
	case "uint8":
		val, err := r.board.client.ReadUInt8(r.offset, r.pinType)
		return board.AnalogValue{Value: int(val)}, err
	case "uint16":
		val, err := r.board.client.ReadUInt16(r.offset, r.pinType)
		return board.AnalogValue{Value: int(val)}, err
	case "int16":
		val, err := r.board.client.ReadInt16(r.offset, r.pinType)
		return board.AnalogValue{Value: int(val)}, err
	case "uint32":
		val, err := r.board.client.ReadUInt32(r.offset, r.pinType)
		return board.AnalogValue{Value: int(val)}, err
	case "int32":
		val, err := r.board.client.ReadInt32(r.offset, r.pinType)
		return board.AnalogValue{Value: int(val)}, err
	case "int64":
		val, err := r.board.client.ReadInt64(r.offset, r.pinType)
		return board.AnalogValue{Value: int(val)}, err
	case "uint64":
		val, err := r.board.client.ReadUInt64(r.offset, r.pinType)
		return board.AnalogValue{Value: int(val)}, err
	case "float32":
		val, err := r.board.client.ReadFloat32(r.offset, r.pinType)
		return board.AnalogValue{Value: int(val)}, err
	case "float64":
		val, err := r.board.client.ReadFloat64(r.offset, r.pinType)
		return board.AnalogValue{Value: int(val)}, err
	default:
		return board.AnalogValue{}, errors.New("invalid data type")
	}
}

func (r *ModbusAnalogPin) Write(ctx context.Context, value int, extra map[string]interface{}) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	switch r.dataType {
	case "int16":
		return r.board.client.WriteInt16(r.offset, int16(value))
	case "uint16":
		return r.board.client.WriteUInt16(r.offset, uint16(value))
	case "int32":
		return r.board.client.WriteInt32(r.offset, int32(value))
	case "uint32":
		return r.board.client.WriteUInt32(r.offset, uint32(value))
	case "int64":
		return r.board.client.WriteInt64(r.offset, int64(value))
	case "uint64":
		return r.board.client.WriteUInt64(r.offset, uint64(value))
	case "float32":
		return r.board.client.WriteFloat32(r.offset, float32(value))
	case "float64":
		return r.board.client.WriteFloat64(r.offset, float64(value))
	default:
		return errors.New("invalid data type")
	}
}

func NewModbusAnalogPin(board *ModbusBoard, conf modbusAnalogPinCloudConfig) (*ModbusAnalogPin, error) {
	// Why do we need a board here? because we need to reinitialize the modbus client if there is an error. this can probably be done better, but here we are
	var t common.RegisterType
	switch common.NewPinType(conf.PinType) {
	case common.InputPin:
		t = common.InputRegister
	case common.OutputPin:
		t = common.HoldingRegister
	default:
		return nil, common.ErrInvalidPinType
	}
	return &ModbusAnalogPin{
		mu:       &board.mu,
		offset:   uint16(conf.Offset),
		dataType: conf.DataType,
		pinType:  t,
		board:    board,
	}, nil
}
