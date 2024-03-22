# viam-modbus

The viam-modbus module currently supports TCP connections to a modbus device.

## Configuring your modbus connection

After creating a new modbus board component, several attributes can be added to specify certain configurations of your board:

```json
{
  "modbus": {
    "url": "<address>:<port>",
    "word_order": "<low/high>",
    "endianness": "<little/big>",
    "timeout_ms": "<timeout_value>"
  },
  "gpio_pins": [
    {
      "pin_type": "<input/output>", 
      "name": "<name>",
      "offset": "<address_offset_value>"
    }
  ],
  "analog_pins": [
    {
      "pin_type": "<input/output>", 
      "name": "<name>",
      "offset": "<address_offset_value>",
      "data_type": "<data_format_type>"
    }
  ],
}
```

GPIO/Analog pins are defined using a pin_type (input/output) a name (any string), and an offset value which represents the address offset from the base address. Analog pins have an additional parameter, "data_type", that represents the data format information (EX: uint8, uint16, ...).

The modbus defines the connection to the device. This includes the URL and port that information will be sent/received through, the word_order and endianness for encoding/decoding, and a timeout before a message will return an error.


## TODO:
  - modbus RTU
