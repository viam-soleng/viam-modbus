# Viam Modbus Integration Module

This repository contains the `board` and `sensor` package, these modules abstract away a modbus server as a board component, and can query the state of that server and return sensor readings.

Available via the [Viam Registry](https://app.viam.com/module/viam-soleng/viam-modbus)! -> Currently for linux/arm64 others will follow soon.

## Description

The Modbus TCP module enables seamless communication between devices by acting as a client that queries data from a Modbus server, facilitating real-time monitoring and control over networked industrial equipment. It allows for reading and writing of coils or registers on the server, enabling efficient data exchange and operational command execution. This module is essential for integrating diverse devices, ensuring interoperability and enhancing automation in industrial environments.

![alt text](media/architecture.png "Modbus Integration (Server / Client) Architecture")

## Features

- **Data Acquisition:** Enables the reading of sensor data from a Modbus Server using a TCP connection.
- **Device Control:** Enables the reading of writing of coils and registers onto a Modbus Server using a TCP connection.
- **Configurable Parameters:** Offers customization options for the device address, word_order, endianness, timeouts, pin types, data types and more.


## How To
This module provides 2 Viam component types: [Board](https://docs.viam.com/components/board/) and [Sensor](https://docs.viam.com/components/sensor/). 

The `board` implementation implements as much of the Viam `board` interface as applicable. This enables you to tread a PLC as a Viam `board` and access the I/O just as you would with a Revolution Pi or other board in Viam.

The `sensor` implementation provides a superset of functionality to the `board` interface except it is read-only. Since a `sensor` is not bound by the `board` interface, it allows for reading arbitrary tags in the PLC that are exposed over Modbus.

### Common Configuration

Whether you choose the `sensor` or `board` implementation, there are some common configuration options required to use these components. These common options are set via the [`modbus` field](/common/modbus_tcp.go#L13) in the configuration.
```json
{
  "modbus":{
    "url": "tcp://x.x.x.x:yyy",
    "timeout_ms": 10000,
    "endianness": "big",
    "word_order": "low"
  }
}
```
|Field|Type|Acceptable Values|Description|
|-----|----|-----------------|-----------|
|`url`|string|(tcp|udp)://<IP>:<PORT>|This is the IP address and port of the Modbus server (slave) you wish to conntect to|
|`timeout_ms`|int32|Any positive integer|This value should be non-0. A timeout of 0 or omitting the field will result in no timeout and could mask configuration/communication errors|
|`endianness`|string|big,little|Specif the [endianness](https://en.wikipedia.org/wiki/Endianness) of the PLC. Setting the wrong value here will result in garbled or incorrect values|
|`word_order`|string|low,high|This value dictates the ordering of words in the PLC. Think byte ordering, vs the bit ordering of endianness|

### Board Component
The board allows for reading/writing: `coils`, `discrete_inputs` and `input_registers`.
|Modbus Type|PLC Type|Read|Write|
|-|-|-|-|
|`coils`|Digital Outputs|Y|Y|
|`discrete_inputs`|Digital Inputs|Y|N|
|`input_registers`|Analog Inputs/Outputs|Y|Y|

#### Configuration
The `board` configuration extends the [common configuration](#common-configuration) to add `gpio-pins` and `analog-pins`.
##### GPIO Pins
GPIO Pins are configured under the `gpio-pins` field. This is an array of [`ModbusGpioPinCloudConfig`](/modbus_tcp_board/config.go#L18) objects. To configure a gpio pin, you need to specify 3 fields: `name`, `offset`, and `pin_type`.

|Field|Type|Acceptable Values|Description|
|-----|----|-----------------|-----------|
|`offset`|int32|Any positive integer|The offset inside the Modbus address space. Some modbus implementations start at 0, some start 1, so be wary of off-by-one errors|
|`name`|string|Any string|This is a pretty name to be assigned to the name. We suggest being consistent with the PLCs internal tag naming convention to avoid any confusion|
|`pin_type`|string|input,output|Specifies whether the pin is an input or an output. Both can be read, reading an output simply reads the state of the command while writing to an input will do nothing|

###### GPIO Pin Example
For an input at offset 2 with the name "DI.0.2" in the PLC tag database
```json
{
  "offset": 2,
  "name": "DI.0.2",
  "pin_type": "input"
}
```
For an output at offset 4 with the name "DO.0.4" in the PLC tag database
```json
{
  "offset": 4,
  "name": "DO.0.4",
  "pin_type": "output"
}
```

##### Analog Pins
Analog pins are configured under the `analog-pins` field. This is an array of [`ModbusAnalogPinCloudConfig`](/modbus_tcp_board/config.go#L24). To configure an analog pin, you need to specify 4 fields: `offset`, `name`, `pin_type`, and `data_type`. Unlike GPIO pins, an analog pin can have several different data types depending on the input.
|Field|Type|Acceptable Values|Description|
|-----|----|-----------------|-----------|
|`offset`|int32|Any positive integer|The offset inside the Modbus address space. Some modbus implementations start at 0, some start 1, so be wary of off-by-one errors|
|`name`|string|Any string|This is a pretty name to be assigned to the name. We suggest being consistent with the PLCs internal tag naming convention to avoid any confusion|
|`pin_type`|string|input,output|Specifies whether the pin is an input or an output. Both can be read, reading an output simply reads the state of the command while writing to an input will do nothing|
|`data_type`|string|int16,uin16,int32,uint32,float32,float64|Specifies the type of the data to be read/written. Some analogs are simple uint16 values, other are more complex like float32 values|

###### Analog Pin Example
For an analog input that is a float32 at offset 2 with the name "DAI.0.2" in the PLC tag database
```json
{
  "offset": 2,
  "name": "DAI.0.2",
  "pin_type": "input",
  "data_type": "float32"
}
```
For an analog output that is an int16 at offset 4 with the name "DAO.0.4" in the PLC tag database
```json
{
  "offset": 4,
  "name": "DO.0.4",
  "pin_type": "output",
  "data_type": "int16"
}
```

##### Full Example

Putting all this together, we get a sample config of:

Sample Configuration Attributes for a Board Component:
```json
{
  "gpio_pins": [
    {
      "offset": 2,
      "name": "DI.0.2",
      "pin_type": "input"
    },
    {
      "offset": 4,
      "name": "DO.0.4",
      "pin_type": "output"
    }
  ],
  "analog_pins": [
    {
      "offset": 2,
      "name": "DAI.0.2",
      "pin_type": "input",
      "data_type": "float32"
    },
    {
      "offset": 4,
      "name": "DO.0.4",
      "pin_type": "output",
      "data_type": "int16"
    }
  ],
  "modbus": {
    "url": "tcp://10.1.12.124:502",
    "word_order": "low",
    "endianness": "big",
    "timeout_ms": 10000
  }
}
```

### Sensor Component
Since the sensor component is not bound by the `board` interface, it is able to read a wider variety of data types and PLC tags than its `board` counterpart.

Sample Configuration Attributes for a Sensor Component:
```json
{
  "blocks": [
    {
      "length": 1,
      "name": "potentiometer",
      "offset": 0,
      "type": "input_registers"
    },
    {
      "length": 4,
      "name": "switches_buttons",
      "offset": 4,
      "type": "discrete_inputs"
    },
    {
      "length": 4,
      "name": "lights",
      "offset": 0,
      "type": "coils"
    },
    {
      "offset": 0,
      "type": "holding_registers",
      "length": 1,
      "name": "voltageDial"
    }
  ],
  "modbus": {
    "timeout_ms": 10000,
    "url": "tcp://10.1.12.124:502",
    "word_order": "low",
    "endianness": "big"
  }
}
```

TODO:
  - modbus TCP
  - modbus RTU
