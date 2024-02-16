package modbus_tcp_board

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/simonvetter/modbus"
	commonpb "go.viam.com/api/common/v1"
	pb "go.viam.com/api/component/board/v1"
	"go.viam.com/rdk/components/board"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"

	"viam-modbus/utils"
)

var Model = resource.NewModel("viam-soleng", "board", "modbus-tcp")

func init() {
	resource.RegisterComponent(
		board.API,
		Model,
		resource.Registration[board.Board, *ModbusTcpBoardConfig]{
			Constructor: NewModbusTcpSensor,
		},
	)
}

func NewModbusTcpSensor(ctx context.Context, deps resource.Dependencies, conf resource.Config, logger logging.Logger) (board.Board, error) {
	logger.Infof("Starting Modbus TCP Sensor Component %v", utils.Version)
	c, cancelFunc := context.WithCancel(context.Background())
	b := ModbusTcpBoard{
		Named:      conf.ResourceName().AsNamed(),
		logger:     logger,
		cancelFunc: cancelFunc,
		ctx:        c,
	}

	if err := b.Reconfigure(ctx, deps, conf); err != nil {
		return nil, err
	}
	return &b, nil
}

type ModbusTcpBoard struct {
	resource.Named
	mu         sync.RWMutex
	logger     logging.Logger
	cancelFunc context.CancelFunc
	ctx        context.Context
	client     *modbus.ModbusClient
	uri        string
	timeout    time.Duration
	endianness modbus.Endianness
	wordOrder  modbus.WordOrder
	gpioPins   map[string]*ModbusGpioPin
	analogPins map[string]*ModbusAnlogPin
}

// AnalogReaderByName implements board.Board.
func (r *ModbusTcpBoard) AnalogReaderByName(name string) (board.AnalogReader, bool) {
	if pin, ok := r.analogPins[name]; ok {
		return pin, false
	}
	return nil, false
}

// AnalogReaderNames implements board.Board.
func (*ModbusTcpBoard) AnalogReaderNames() []string {
	return nil
}

// DigitalInterruptByName implements board.Board.
func (*ModbusTcpBoard) DigitalInterruptByName(name string) (board.DigitalInterrupt, bool) {
	return nil, false
}

// DigitalInterruptNames implements board.Board.
func (*ModbusTcpBoard) DigitalInterruptNames() []string {
	return nil
}

// GPIOPinByName implements board.Board.
func (r *ModbusTcpBoard) GPIOPinByName(name string) (board.GPIOPin, error) {
	if pin, ok := r.gpioPins[name]; ok {
		return pin, nil
	}
	return nil, errors.New("pin not found")
}

// SetPowerMode implements board.Board.
func (*ModbusTcpBoard) SetPowerMode(ctx context.Context, mode pb.PowerMode, duration *time.Duration) error {
	return errors.ErrUnsupported
}

// Status implements board.Board.
func (*ModbusTcpBoard) Status(ctx context.Context, extra map[string]interface{}) (*commonpb.BoardStatus, error) {
	return nil, nil
}

// WriteAnalog implements board.Board.
func (*ModbusTcpBoard) WriteAnalog(ctx context.Context, pin string, value int32, extra map[string]interface{}) error {
	return errors.ErrUnsupported
}

func (r *ModbusTcpBoard) Close(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logger.Info("Closing Modbus TCP Sensor Component")
	r.cancelFunc()
	if r.client != nil {
		err := r.client.Close()
		if err != nil {
			r.logger.Errorf("Failed to close modbus client: %#v", err)
		}
	}
	return nil
}

// DoCommand implements resource.Resource.
func (*ModbusTcpBoard) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{"ok": 1}, nil
}

// Reconfigure implements resource.Resource.
func (r *ModbusTcpBoard) Reconfigure(ctx context.Context, deps resource.Dependencies, conf resource.Config) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logger.Debug("Reconfiguring Modbus TCP Sensor Component")

	newConf, err := resource.NativeConfig[*ModbusTcpBoardConfig](conf)
	if err != nil {
		return err
	}

	// In case the module has changed name
	r.Named = conf.ResourceName().AsNamed()

	return r.reconfigure(newConf, deps)
}

func (r *ModbusTcpBoard) reconfigure(newConf *ModbusTcpBoardConfig, deps resource.Dependencies) error {
	r.logger.Infof("Reconfiguring Modbus TCP Sensor Component with %v", newConf)
	if r.client != nil {
		err := r.client.Close()
		if err != nil {
			r.logger.Errorf("Failed to close modbus client: %#v", err)
			// TODO: should we exit here?
		}
	}
	r.uri = newConf.Url
	r.timeout = time.Millisecond * time.Duration(newConf.Timeout)

	switch newConf.Endianness {
	case "big":
		r.endianness = modbus.BIG_ENDIAN
	case "little":
		r.endianness = modbus.LITTLE_ENDIAN
	default:
		return errors.New("invalid endianness")
	}

	switch newConf.WordOrder {
	case "high":
		r.wordOrder = modbus.HIGH_WORD_FIRST
	case "low":
		r.wordOrder = modbus.LOW_WORD_FIRST
	default:
		return errors.New("invalid word order")
	}

	err := r.initializeModbusClient()
	if err != nil {
		return err
	}
	r.gpioPins = map[string]*ModbusGpioPin{}
	for _, pinConf := range newConf.GpioPins {
		pin := NewModbusGpioPin(r, pinConf)
		r.gpioPins[pinConf.Name] = pin
	}

	r.analogPins = map[string]*ModbusAnlogPin{}
	for _, pinConf := range newConf.AnalogPins {
		pin, err := NewModbusAnalogPin(r, pinConf)
		if err != nil {
			return err
		}
		r.analogPins[pinConf.Name] = pin
	}
	return nil
}

func (r *ModbusTcpBoard) initializeModbusClient() error {
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
	client.SetEncoding(r.endianness, r.wordOrder)
	r.client = client
	return nil
}
