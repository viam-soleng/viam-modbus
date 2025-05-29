package modbus_connection

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/simonvetter/modbus"
	"go.viam.com/rdk/components/generic"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"

	"viam-modbus/common"
	"viam-modbus/modbus_sensor"
	"viam-modbus/utils"
)

var Model = resource.NewModel("viam-soleng", "generic", "modbus-connection")

// Registers the generic connection model
func init() {
	resource.RegisterComponent(
		generic.API,
		Model,
		resource.Registration[resource.Resource, *ModbusConnectionConfig]{
			Constructor: newModbusConnectionModel,
		},
	)
}

// Creates a new modbus connection instance
func newModbusConnectionModel(ctx context.Context, deps resource.Dependencies, rawConf resource.Config, logger logging.Logger) (resource.Resource, error) {
	logger.Infof("Starting Modbus Connection Component %v", utils.Version)

	c, cancelFunc := context.WithCancel(context.Background())
	b := ModbusConnection{
		Named:      rawConf.ResourceName().AsNamed(),
		logger:     logger,
		cancelFunc: cancelFunc,
		ctx:        c,
	}

	if err := b.Reconfigure(ctx, deps, rawConf); err != nil {
		return nil, err
	}
	return &b, nil
}

type ModbusConnection struct {
	resource.Named
	mu         sync.RWMutex
	logger     logging.Logger
	cancelFunc context.CancelFunc
	ctx        context.Context
	client     *common.ViamModbusClient
}

// Validate ensures all parts of the config are valid and important fields exist.
// Returns implicit dependencies based on the config.
// The path is the JSON path in your robot's config (not the `Config` struct) to the
// resource being validated; e.g. "components.0".
func (r *ModbusConnection) Validate(path string) ([]string, []string, error) {
	// Add config validation code here
	return nil, nil, nil
}

// Closes the modbus connection instance
func (r *ModbusConnection) Close(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logger.Info("Closing Modbus Connection Component")
	r.cancelFunc()
	if r.client != nil {
		err := r.client.Close()
		if err != nil {
			r.logger.Errorf("Failed to close modbus client: %#v", err)
		}
	}
	return nil
}

// DoCommand() receives a stringified ModbusBlocks[] from a modbus_sensor Readings()
func (r *ModbusConnection) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	//r.logger.Infof("ModbusConnection DoCommand() starting with %v", cmd["blocks"])
	var Blocks []modbus_sensor.ModbusBlocks
	err := json.Unmarshal([]byte(cmd["blocks"].(string)), &Blocks)
	if err != nil {
		r.logger.Infof("Error unmarshaling JSON:", err)
		return nil, err
	}

	results := map[string]interface{}{}
	for _, block := range Blocks {
		switch block.Type {
		case "coils":
			b, err := r.client.ReadCoils(uint16(block.Offset), uint16(block.Length))
			if err != nil {
				return nil, err
			}
			writeBoolArrayToOutput(b, block, results)
		case "discrete_inputs":
			b, err := r.client.ReadDiscreteInputs(uint16(block.Offset), uint16(block.Length))
			if err != nil {
				return nil, err
			}
			writeBoolArrayToOutput(b, block, results)
		case "holding_registers":
			b, err := r.client.ReadHoldingRegisters(uint16(block.Offset), uint16(block.Length))
			if err != nil {
				return nil, err
			}
			writeUInt16ArrayToOutput(b, block, results)
		case "input_registers":
			b, err := r.client.ReadInputRegisters(uint16(block.Offset), uint16(block.Length))
			if err != nil {
				return nil, err
			}
			writeUInt16ArrayToOutput(b, block, results)
		case "bytes":
			b, e := r.client.ReadBytes(uint16(block.Offset), uint16(block.Length), modbus.HOLDING_REGISTER)
			if e != nil {
				return nil, e
			}
			writeByteArrayToOutput(b, block, results)
		case "rawBytes":
			b, e := r.client.ReadRawBytes(uint16(block.Offset), uint16(block.Length), modbus.HOLDING_REGISTER)
			if e != nil {
				return nil, e
			}
			writeByteArrayToOutput(b, block, results)
		case "uint8":
			b, e := r.client.ReadUInt8(uint16(block.Offset), modbus.HOLDING_REGISTER)
			if e != nil {
				return nil, e
			}
			results[block.Name] = int32(b)
		case "int16":
			b, e := r.client.ReadInt16(uint16(block.Offset), modbus.HOLDING_REGISTER)
			if e != nil {
				return nil, e
			}
			results[block.Name] = int32(b)
		case "uint16":
			b, e := r.client.ReadUInt16(uint16(block.Offset), modbus.HOLDING_REGISTER)
			if e != nil {
				return nil, e
			}
			results[block.Name] = int32(b)
		case "int32":
			b, e := r.client.ReadInt32(uint16(block.Offset), modbus.HOLDING_REGISTER)
			if e != nil {
				return nil, e
			}
			results[block.Name] = b
		case "uint32":
			b, e := r.client.ReadUInt32(uint16(block.Offset), modbus.HOLDING_REGISTER)
			if e != nil {
				return nil, e
			}
			results[block.Name] = b
		case "float32":
			b, e := r.client.ReadFloat32(uint16(block.Offset), modbus.HOLDING_REGISTER)
			if e != nil {
				return nil, e
			}
			results[block.Name] = b
		case "float64":
			b, e := r.client.ReadFloat64(uint16(block.Offset), modbus.HOLDING_REGISTER)
			if e != nil {
				return nil, e
			}
			results[block.Name] = b
		default:
			results[block.Name] = "unsupported type"
		}
	}
	//r.logger.Infof("ModbusConnection DoCommand() finished with %v", results)
	return results, nil
}

// Configures the modbus sensor instance
func (r *ModbusConnection) Reconfigure(ctx context.Context, deps resource.Dependencies, conf resource.Config) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logger.Debug("Reconfiguring Modbus Connection Component")

	newConf, err := resource.NativeConfig[*ModbusConnectionConfig](conf)
	if err != nil {
		return err
	}

	// In case the module has changed name
	r.Named = conf.ResourceName().AsNamed()

	return r.reconfigure(newConf, deps)
}

// func: reconfigure() - close any open connections, open a new connection to the Modbus server
func (r *ModbusConnection) reconfigure(newConf *ModbusConnectionConfig, _ resource.Dependencies) error {
	r.logger.Infof("Reconfiguring Modbus Connection Component with %v", newConf)
	if r.client != nil {
		err := r.client.Close()
		if err != nil {
			r.logger.Errorf("Failed to close modbus client: %#v", err)
			// TODO: should we exit here?
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

	if newConf.Modbus.URL == "" {
		return fmt.Errorf("No Modbus server URL defined.")
	}

	clientConfig := modbus.ClientConfiguration{
		URL:      newConf.Modbus.URL,
		Speed:    newConf.Modbus.Speed,
		DataBits: newConf.Modbus.DataBits,
		Parity:   newConf.Modbus.Parity,
		StopBits: newConf.Modbus.StopBits,
		Timeout:  timeout,
		// TODO: To be implemented
		//TLSClientCert: tlsClientCert,
		//TLSRootCAs:    tlsRootCAs,
	}
	client, err := common.NewModbusClient(r.logger, endianness, wordOrder, clientConfig)
	if err != nil {
		return err
	}
	r.client = client

	return nil
}

// Helper function - conversion
func writeBoolArrayToOutput(b []bool, block modbus_sensor.ModbusBlocks, results map[string]interface{}) {
	// only rename block Name with "_0", "_1" if there are more than one in this array
	if len(b) > 1 {
		for i, v := range b {
			field_name := block.Name + "_" + fmt.Sprint(i)
			results[field_name] = v
		}
	} else {
		results[block.Name] = b[0]
	}
}

// Helper function - conversion
func writeUInt16ArrayToOutput(b []uint16, block modbus_sensor.ModbusBlocks, results map[string]interface{}) {
	// only rename block Name with "_0", "_1" if there are more than one in this array
	if len(b) > 1 {
		for i, v := range b {
			field_name := block.Name + "_" + fmt.Sprint(i)
			//results[field_name] = strconv.Itoa(int(v))
			results[field_name] = int(v)
		}
	} else {
		//results[block.Name] = strconv.Itoa(int(b[0]))
		results[block.Name] = int(b[0])
	}
}

// Helper function - conversion
func writeByteArrayToOutput(b []byte, block modbus_sensor.ModbusBlocks, results map[string]interface{}) {
	results[block.Name] = hex.EncodeToString(b)
}
