package modbus_tcp_board

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"time"

	"viam-modbus/common"
	"viam-modbus/utils"

	pb "go.viam.com/api/component/board/v1"
	"go.viam.com/rdk/components/board"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
)

var Model = resource.NewModel("viam-soleng", "board", "modbus-tcp")

func init() {
	resource.RegisterComponent(
		board.API,
		Model,
		resource.Registration[board.Board, *ModbusTcpBoardCloudConfig]{
			Constructor: NewModbusTcpBoard,
		},
	)
}

func NewModbusTcpBoard(ctx context.Context, deps resource.Dependencies, conf resource.Config, logger logging.Logger) (board.Board, error) {
	logger.Infof("Starting Modbus TCP Board Component %v", utils.Version)
	c, cancelFunc := context.WithCancel(context.Background())
	b := ModbusTcpBoard{
		Named:      conf.ResourceName().AsNamed(),
		logger:     logger,
		cancelFunc: cancelFunc,
		ctx:        c,
	}

	if err := b.Reconfigure(ctx, deps, conf); err != nil {
		logger.Errorf("Failed to start Modbus TCP Board Component %v", err)
		return nil, err
	}
	logger.Info("Modbus TCP Board Component started successfully")
	return &b, nil
}

type ModbusTcpBoard struct {
	resource.Named
	client     *common.ModbusTcpClient
	mu         sync.RWMutex
	logger     logging.Logger
	cancelFunc context.CancelFunc
	ctx        context.Context

	gpioPins   map[string]*ModbusGpioPin
	analogPins map[string]*ModbusAnalogPin
}

func (r *ModbusTcpBoard) getAnalogPin(name string) (*ModbusAnalogPin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if pin, ok := r.analogPins[name]; ok {
		return pin, nil
	}
	return nil, errors.New("pin not found")
}

// AnalogReaderByName implements board.Board.
func (r *ModbusTcpBoard) AnalogByName(name string) (board.Analog, error) {
	pin, err := r.getAnalogPin(name)
	if err == nil {
		return pin, nil
	}
	return nil, err
}

// WriteAnalog implements board.Board.
func (r *ModbusTcpBoard) WriteAnalog(ctx context.Context, name string, value int32, extra map[string]interface{}) error {
	if pin, err := r.getAnalogPin(name); err == nil {
		return pin.Write(ctx, int(value), extra)
	}
	return errors.New("pin not found")
}

// AnalogReaderNames implements board.Board.
func (*ModbusTcpBoard) AnalogNames() []string {
	return nil
}

// DigitalInterruptByName implements board.Board.
func (*ModbusTcpBoard) DigitalInterruptByName(name string) (board.DigitalInterrupt, error) {
	return nil, errors.ErrUnsupported
}

// DigitalInterruptNames implements board.Board.
func (*ModbusTcpBoard) DigitalInterruptNames() []string {
	return nil
}

// GPIOPinByName implements board.Board.
func (r *ModbusTcpBoard) GPIOPinByName(name string) (board.GPIOPin, error) {
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
func (*ModbusTcpBoard) SetPowerMode(ctx context.Context, mode pb.PowerMode, duration *time.Duration) error {
	return errors.ErrUnsupported
}

// StreamTicks starts a stream of digital interrupt ticks.
func (r *ModbusTcpBoard) StreamTicks(ctx context.Context, interrupts []board.DigitalInterrupt, ch chan board.Tick,
	extra map[string]interface{},
) error {
	return errors.ErrUnsupported
}

func (r *ModbusTcpBoard) Close(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logger.Info("Closing Modbus TCP Board Component")
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
	r.logger.Debug("Parsing new configuration for Modbus TCP Board")

	newConf, err := resource.NativeConfig[*ModbusTcpBoardCloudConfig](conf)
	if err != nil {
		return err
	}

	// In case the module has changed name
	r.Named = conf.ResourceName().AsNamed()

	return r.reconfigure(newConf, deps)
}

func (r *ModbusTcpBoard) reconfigure(newConf *ModbusTcpBoardCloudConfig, deps resource.Dependencies) error {
	r.logger.Infof("Reconfiguring Modbus TCP Board Component with %v", newConf)
	if r.client != nil {
		err := r.client.Close()
		if err != nil {
			r.logger.Errorf("Failed to close modbus client: %#v", err)
			// TODO: should we exit here?
			os.Exit(1)
		}
	}

	endianness, err := common.GetEndianness(newConf.Modbus.Endianness)
	if err != nil {
		return err
	}

	wordOrder, err := common.GetWordOrder(newConf.Modbus.WordOrder)
	if err != nil {
		return err
	}

	timeout := time.Millisecond * time.Duration(newConf.Modbus.Timeout)
	client, err := common.NewModbusTcpClient(r.logger, newConf.Modbus.Url, timeout, endianness, wordOrder)
	if err != nil {
		r.logger.Errorf("Failed to initialize modbus client: %#v", err)
		return err
	}
	r.client = client
	r.logger.Debugf("Initialized modbus client")
	r.gpioPins = map[string]*ModbusGpioPin{}
	for _, pinConf := range newConf.GpioPins {
		r.logger.Debugf("Creating GPIO pin: %v", pinConf.Name)
		pinType := common.NewPinType(pinConf.PinType)
		if pinType == common.UNKNOWN {
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
	r.logger.Debug("Done initializing Modbus TCP Board")
	return nil
}
