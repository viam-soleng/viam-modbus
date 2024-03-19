package common

import (
	"errors"

	"github.com/simonvetter/modbus"
)

type PinType string

const (
	OUTPUT_PIN PinType = "output"
	INPUT_PIN  PinType = "input"
	UNKNOWN    PinType = "unknown"
)

var ErrInvalidPinType = errors.New("invalid pin type")
var ErrSetInputPin = errors.New("cannot set input pin")
var ErrRetriesExhausted = errors.New("retries exhausted")

func NewPinType(s string) PinType {
	switch s {
	case "output":
		return OUTPUT_PIN
	case "input":
		return INPUT_PIN
	default:
		return UNKNOWN
	}
}

func GetEndianness(s string) (modbus.Endianness, error) {
	switch s {
	case "big":
		return modbus.BIG_ENDIAN, nil
	case "little":
		return modbus.LITTLE_ENDIAN, nil
	default:
		return 0, errors.New("invalid endianness")
	}
}

func GetWordOrder(s string) (modbus.WordOrder, error) {
	switch s {
	case "high":
		return modbus.HIGH_WORD_FIRST, nil
	case "low":
		return modbus.LOW_WORD_FIRST, nil
	default:
		return 0, errors.New("invalid word order")
	}
}
