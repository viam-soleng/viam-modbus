// Package board provides a board component that reads data from a modbus server and exposes it as a board in Viam
package board

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	pb "go.viam.com/api/component/board/v1"
	"go.viam.com/rdk/components/board"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"

	"viam-modbus/common"
)

type resourceType = board.Board
type config = *modbusBoardCloudConfig

var API = board.API
var Model = resource.NewModel("viam-soleng", "modbus", "board")

func init() {
	resource.RegisterComponent(
		API,
		Model,
		resource.Registration[resourceType, config]{
			Constructor: NewModbusBoard,
		},
	)
}

func NewModbusBoard(ctx context.Context, deps resource.Dependencies, conf resource.Config, logger logging.Logger) (resourceType, error) {
	logger.Infof("Starting Modbus Board Component %v", common.Version)
	c, cancelFunc := context.WithCancel(context.Background())
	b := ModbusBoard{
		Named:      conf.ResourceName().AsNamed(),
		logger:     logger,
		cancelFunc: cancelFunc,
		ctx:        c,
	}

	if err := b.Reconfigure(ctx, deps, conf); err != nil {
		logger.Errorf("Failed to start Modbus Board Component %v", err)
		return nil, err
	}
	logger.Info("Modbus Board Component started successfully")
	return &b, nil
}

type ModbusBoard struct {
	resource.Named
	client     *common.ViamModbusClient
	mu         sync.RWMutex
	logger     logging.Logger
	cancelFunc context.CancelFunc
	ctx        context.Context

	gpioPins   map[string]*ModbusGpioPin
	analogPins map[string]*ModbusAnalogPin
}

func (r *ModbusBoard) getAnalogPin(name string) (*ModbusAnalogPin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if pin, ok := r.analogPins[name]; ok {
		return pin, nil
	}
	return nil, errors.New("pin not found")
}

// AnalogByName implements board.Board.
func (r *ModbusBoard) AnalogByName(name string) (board.Analog, error) {
	pin, err := r.getAnalogPin(name)
	if err != nil {
		return nil, err
	}
	return pin, nil
}

// AnalogNames implements board.Board.
func (*ModbusBoard) AnalogNames() []string {
	return nil
}

// DigitalInterruptByName implements board.Board.
func (*ModbusBoard) DigitalInterruptByName(name string) (board.DigitalInterrupt, error) {
	return nil, errors.ErrUnsupported
}

// DigitalInterruptNames implements board.Board.
func (*ModbusBoard) DigitalInterruptNames() []string {
	return nil
}

// GPIOPinByName implements board.Board.
func (r *ModbusBoard) GPIOPinByName(name string) (board.GPIOPin, error) {
	r.logger.Debugf("Getting GPIO pin by name: %v %T", name, name)
	// Hack to fix data capture bug
	if strings.HasPrefix(name, "[") {
		lenToTrim := len("[type.googleapis.com/google.protobuf.StringValue]:{value:\"")
		s := name[lenToTrim:]
		s = s[:len(s)-2]
		name = s
	}
	if pin, ok := r.gpioPins[name]; ok {
		return pin, nil
	}
	return nil, errors.New("pin not found")
}

// SetPowerMode implements board.Board.
func (*ModbusBoard) SetPowerMode(ctx context.Context, mode pb.PowerMode, duration *time.Duration) error {
	return errors.ErrUnsupported
}

// StreamTicks implements board.Board.
func (*ModbusBoard) StreamTicks(ctx context.Context, interrupts []board.DigitalInterrupt, ch chan board.Tick, extra map[string]interface{}) error {
	return errors.ErrUnsupported
}

func (r *ModbusBoard) Close(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logger.Info("Closing Modbus Board Component")
	r.cancelFunc()
	// if r.client != nil {
	// 	err := r.client.Close()
	// 	if err != nil {
	// 		r.logger.Errorf("Failed to close modbus client: %#v", err)
	// 	}
	// }
	return nil
}

// DoCommand implements resource.Resource.
func (*ModbusBoard) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{"ok": 1}, nil
}

// Reconfigure implements resource.Resource.
func (r *ModbusBoard) Reconfigure(ctx context.Context, deps resource.Dependencies, conf resource.Config) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logger.Debug("Parsing new configuration for Modbus Board")

	newConf, err := resource.NativeConfig[*modbusBoardCloudConfig](conf)
	if err != nil {
		return err
	}

	// In case the module has changed name
	r.Named = conf.ResourceName().AsNamed()

	return r.reconfigure(newConf, deps)
}

func (r *ModbusBoard) reconfigure(newConf *modbusBoardCloudConfig, _ resource.Dependencies) error {
	r.logger.Infof("Reconfiguring Modbus Board Component with %v", newConf)

	client, err := common.NewModbusClient(r.logger, newConf.Modbus)
	if err != nil {
		return err
	}
	r.client = client
	r.logger.Debugf("Initialized modbus client")
	r.gpioPins = map[string]*ModbusGpioPin{}
	for _, pinConf := range newConf.GpioPins {
		r.logger.Debugf("Creating GPIO pin: %v", pinConf.Name)
		pinType := common.NewPinType(pinConf.PinType)
		if pinType == common.Unknown {
			return common.ErrInvalidPinType
		}
		pin := NewModbusGpioPin(r, uint16(pinConf.Offset), pinType)
		r.gpioPins[pinConf.Name] = pin
	}
	r.logger.Debug("Initialized GPIO pins")

	r.analogPins = map[string]*ModbusAnalogPin{}
	for _, pinConf := range newConf.AnalogPins {
		r.logger.Debugf("Creating Analog pin: %v", pinConf.Name)
		pin, err := NewModbusAnalogPin(r, pinConf)
		if err != nil {
			return err
		}
		r.analogPins[pinConf.Name] = pin
	}
	r.logger.Debug("Initialized Analog pins")
	r.logger.Debug("Done initializing Modbus Board")
	return nil
}
