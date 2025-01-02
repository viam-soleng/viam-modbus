# Viam Modbus Client Module

This repository contains the `board`(r/w) and `sensor`(r) package, these modules abstract away a modbus server as a board component, and can query the state of that server and return sensor readings.

Available via the [Viam Registry](https://app.viam.com/module/viam-soleng/viam-modbus)!

## Description

The Viam Modbus module enables seamless communication between devices by acting as a client that queries data from a Modbus server, facilitating real-time monitoring and control over connected industrial equipment. It allows for reading and writing of coils or registers on the server, enabling efficient data exchange and operational command execution. This module is essential for integrating diverse devices, ensuring interoperability and enhancing automation in industrial environments.

![alt text](media/architecture.png "Modbus Integration (Server / Client) Architecture")

## Features

- **Data Acquisition:** Enables the reading of sensor data from a Modbus Server.
- **Device Control:** Enables the reading of writing of coils and registers onto a Modbus Server.
- **Configurable Parameters:** Offers customization options for the device address, word_order, endianness, timeouts, pin types, data types and more.

## Component Types

This library includes several Viam Components:
* Board - A Modbus Client that exposes the registers of the Modbus Server according to the Viam Board API.
* Sensor - A Modbus Client that exposes the registers of the Modbus Server via the Viam Sensor API 
* Client Bridge - A component that runs 2 or more Modbus servers with a shared set of registers enabling bridging of Modbus Clients running on different physical mediums

## Transport Types

There are 2 basic Modbus transport types, "Network" (TCP/UDP) and "Serial" (RTU/ASCII).

### Serial Attributes
To configure any Modbus Serial connection simply set the `serial_config` field, with the following fields:
| Setting | Data Type | Inclusion | Valid Values | Description |
| ------- | --------- | --------- | ------------ | ----------- |
|`server_id`|uint16|**Required**| 1-255 | The address of the server |
|`speed`|uint|**Required**|Any valid baud rate| The baud rate to use for the connection|
|`data_bits`|uint|**Required**|7, 8|The number of bits that make up a single byte|
|`parity`|string|**Required**|"N", "E", "O"| Specifies how the parity is calculated with **N**one, **E**ven, **O**dd|
|`stop_bits`|uint|**Required**|1, 2| ASCII requires 1 stop bit, RTU requires 1 or 2|
|`rtu`|bool|Optional|true, false| True indicates that this is a Modbus RTU configuration|

### Network Attributes
To configure any Modbus Network connection, there are no extra fields required, simply specify an empty `tcp_config` key.

## Modbus Client Configuration

The Viam modbus client module supports connections over tcp and serial. Which mode is used, depends on the `modbus.url` prefix as explained below.
As with any other Viam module you can apply the configuration to your component into the `Configure` section.
There are two configuration areas. The `modbus`config path applies to both, the sensor and the board component.

### TCP Client Example

```json
  {
    "name": "PLCClient",
    "endpoint": "tcp://192.168.1.124:502",
    "tcp_config": {},
  }
```

**TCP Client Configuration Attributes**

| Name    | Type   | Inclusion    | Description |
| ------- | ------ | ------------ | ----------- |
| `name`  | string | optional     | A simple name for the client|
| `url` | string | **Required** | TCP Config: `"tcp://<ip address>:port"`|

### Serial / RTU Client Example

Add this to your modbus board or sensor component to configure the modbus client to use serial communication.

```json
  {
    "name": "MyPLC",
    "endpoint": "/dev/tty...",
    "serial_config": {...},
  }
```

**Serial Client Configuration Attributes**

| Name    | Type   | Inclusion    | Description |
| ------- | ------ | ------------ | ----------- |
| `name`  | string | optional     | A simple name for the client|
| `endpoint` | string | **Required** | Serial Config: `"//<serial device path>"`|
| `serial_config` | SerialConfig | **Required** | The [Serial Configuration](#serial-configuration) |

### Serial / ASCII Client Example

Add this to your modbus board or sensor component to configure the modbus client to use serial communication.

```json
  {
    "name": "MyPLC",
    "endpoint": "/dev/tty...",
    "serial_config": {...}
  }
```

**Serial Client Configuration Attributes**

| Name    | Type   | Inclusion    | Description |
| ------- | ------ | ------------ | ----------- |
| `name`  | string | optional     | A simple name for the client|
| `endpoint` | string | **Required** | Serial Config: `"//<serial device path>"`|
| `serial_config` | SerialConfig | **Required** | The [Serial Configuration](#serial-configuration) |

## Modbus Sensor Configuration

The modbus sensor component allows you to read and reord modbus register values.

### Sensor Component Configuration Example

```json
{
  "blocks": [
    {
      "length": 1,
      "name": "potentiometer",
      "offset": 0,
      "type": "input_registers"
    },
    {...}
  ],
  "modbus": {...}
}
```

### Sensor Component Block Attributes

| Name    | Type   | Inclusion    | Description |
| ------- | ------ | ------------ | ----------- |
| `name` | string | **Required**| Name of the key for the value being read |
| `type` | string | **Required**| "input_registers" \| "discrete_inputs" \| "coils" \| "holding_registers" |
| `offset` | int | **Required** | Register address decimal|
| `length` | int | **Required** | Number of words to include from register address|

**Modbus Data Model / Register Types**

|Register Type | Access | Size | Features |
| ------- | ------ | ------------ | ----------- |
|Coil (discrete output)	| Read-write | 1 bit | Read/Write on/off value |
|Discrete input	| Read-only | 1 bit	| Read on/off value |
|Input register	| Read-only	| 16 bits (0–65,535) | Read measurements and statuses |
|Holding register |	Read-write | 16 bits (0–65,535) | Read/Write configuration values |

[Modbus on Wikipedia](https://en.wikipedia.org/wiki/Modbus)

## Modbus Board Configuration

Sample Configuration Attributes for a Board Component:
```json
{
  "gpio_pins": [
    {
      "pin_type": "input",
      "name": "DI_01",
      "offset": 4
    },
    {...}
  ],
  "analog_pins": [
    {
      "pin_type": "input",
      "data_type": "uint16",
      "name": "AI_01",
      "offset": 0
    },
    {...}
  ],
  "modbus": {...},
}
```

### Board Component Block Attributes

|Path| Name    | Type   | Inclusion    | Description |
| ------- | ------- | ------ | ------------ | ----------- |
|`gpio_pins`\|`analog_pins`| `name` | string | **Required**| Name of the pin |
|`gpio_pins`\|`analog_pins`| `pin_type` | string | **Required**| "input" \| "output" |
|`gpio_pins`\|`analog_pins`| `offset` | int | **Required** | Register address decimal|
|`analog_pins`| `data_type` | string | **Required** | "uint8" \| "uint16" \| "uint32" \| "uint64" \| "float32" \| "float64" |


## Modbus Server Bridge

Create 2 or more servers, that share the same registers, so that multiple clients, across different transports, can read/write the same data.

```json
{
  "servers": [
    {
      "name": "oven",
      "endpoint": "/dev/ttyUSB0",
      "timeout_ms": 10000,
      "serial_config": {...}
    },
    {
      "endpoint": ":502",
      "timeout_ms": 10000,
      "tcp_config": {...},
      "name": "tcp"
    }
  ],
  "persist_data": true
}
```
### Modbus Server Bridge Attributes
| Setting | Data Type | Inclusion | Valid Values | Description |
| ------- | --------- | --------- | ------------ | ----------- |
|`servers`|ModbusConfig[]|**Required**| - |The list of servers to create|


## TODO:
  - Authentication

## Credits
- RinzlerLabs [Go Modbus Library](https://github.com/rinzlerlabs/gomodbus)
