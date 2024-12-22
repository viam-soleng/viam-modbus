package common

import (
	"errors"
)

var Version = "v0.0.4"

// Change the namespace to match your organization name so it can be uploaded to your private registry
var Namespace = "viam-soleng"

// Change the logger name if you want to customize it
var LoggerName = "viam-modbus"

type PinType string

const (
	OutputPin PinType = "output"
	InputPin  PinType = "input"
	Unknown   PinType = "unknown"
)

var (
	ErrInvalidPinType      = errors.New("invalid pin type")
	ErrSetInputPin         = errors.New("cannot set input pin")
	ErrRetriesExhausted    = errors.New("retries exhausted")
	ErrInvalidRegisterType = errors.New("invalid register type")
)

func NewPinType(s string) PinType {
	switch s {
	case "output":
		return OutputPin
	case "input":
		return InputPin
	default:
		return Unknown
	}
}
