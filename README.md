# Viam Modbus Client Module

This repository contains the `connection`, `board`(r/w) and `sensor`(r) packages, these modules abstract away a modbus server as a component, and can query the state of that server and return sensor readings.

Available via the [Viam Registry](https://app.viam.com/module/viam-soleng/viam-modbus)!

## Description

The Viam Modbus module enables seamless communication between devices by acting as a client that queries data from a Modbus server, facilitating real-time monitoring and control over connected industrial equipment. It allows for reading and writing of coils or registers on the server, enabling efficient data exchange and operational command execution. This module is essential for integrating diverse devices, ensuring interoperability and enhancing automation in industrial environments.

![alt text](media/architecture.png "Modbus Integration (Server / Client) Architecture")

## Features

- **Data Acquisition:** Enables the reading of sensor data from a Modbus Server.
- **Device Control:** Enables the reading of writing of coils and registers onto a Modbus Server.
- **Configurable Parameters:** Offers customization options for the device address, word_order, endianness, timeouts, pin types, data types and more.

## Modbus Connection Client Configuration

> [!NOTE]  
> Serial/RTU client not yet published to registry!

The Viam modbus client module supports connections over tcp and serial. Which mode is used, depends on the `modbus.url` prefix as explained below.
As with any other Viam module you can apply the configuration to your component into the `Configure` section.

Add this to your modbus connection generic component to configure the modbus client to use TCP communication.

### TCP Client Example

```json
  "modbus": {
    "url": "tcp://192.168.1.124:502",
    "word_order": "low",
    "endianness": "big",
    "timeout_ms": 10000
  }
```

#### TCP Client Configuration Attributes

| Name              | Type   | Inclusion    | Description                             |
| ----------------- | ------ | ------------ | --------------------------------------- |
| `url`             | string | **Required** | TCP Config: `"tcp://<ip address>:port"` |
| `timeout_ms`      | string | Optional     | Connection timeout                      |
| `endianness`      | string | Optional     |                                         |
| `word_order`      | string | Optional     |                                         |
| `tls_client_cert` | string | Optional     | Not implemented yet                     |
| `tls_root_cas`    | string | Optional     | Not implemented yet                     |

### Serial / RTU Client Example

Add this to your modbus connection generic component to configure the modbus client to use serial communication.

```json
  "modbus": {
    "url": "rtu:///dev/tty...",
    "speed": 115200,
    "timeout_ms": 10000
  }
```

#### Serial Client Configuration Attributes

| Name         | Type   | Inclusion    | Description                                   |
| ------------ | ------ | ------------ | --------------------------------------------- |
| `url`        | string | **Required** | Serial Config: `"rtu://<serial device path>"` |
| `speed`      | string | **Required** | Bit (bit/s)                                   |
| `data_bits`  | uint   | Optional     |                                               |
| `parity`     | uint   | Optional     |                                               |
| `stop_bits`  | uint   | Optional     |                                               |
| `timeout_ms` | string | Optional     | Connection timeout                            |
| `endianness` | string | Optional     |                                               |
| `word_order` | string | Optional     |                                               |

## Modbus Sensor Configuration

The modbus sensor component allows you to read and record modbus register values. Specify the modbus connection generic component name.

### Sensor Component Configuration Example

```json
{
  "modbus_connection_name": "modbus-connection-server",
  "unit_id": 4,
  "blocks": [
    {
      "length": 1,
      "name": "potentiometer",
      "offset": 0,
      "type": "input_registers"
    },
    {...}
  ]
}
```

### Sensor Component Attributes

| Name                     | Type    | Inclusion    | Description                                   |
| ------------------------ | ------- | ------------ | --------------------------------------------- |
| `modbus_connection_name` | string  | **Required** | Name of the key for the value being read      |
| `unit_id`                | int     | **Optional** | Optionally set the unit id, valid range 0-247 |
| `blocks`                 | []Block | **Required** | Registers etc. to read see below              |

### Sensor Component Block Attributes

| Name      | Type   | Inclusion    | Description                                                              |
| --------- | ------ | ------------ | ------------------------------------------------------------------------ |
| `name`    | string | **Required** | Name of the key for the value being read                                 |
| `type`    | string | **Required** | "input_registers" \| "discrete_inputs" \| "coils" \| "holding_registers" |
| `offset`  | int    | **Required** | Register address decimal                                                 |
| `length`  | int    | **Required** | Number of words to include from register address                         |
| `unit_id` | int    | **Optional** | Set the unit id, valid range 0-247                                       |

#### Modbus Data Model / Register Types

| Register Type          | Access     | Size               | Features                        |
| ---------------------- | ---------- | ------------------ | ------------------------------- |
| Coil (discrete output) | Read-write | 1 bit              | Read/Write on/off value         |
| Discrete input         | Read-only  | 1 bit              | Read on/off value               |
| Input register         | Read-only  | 16 bits (0-65,535) | Read measurements and statuses  |
| Holding register       | Read-write | 16 bits (0-65,535) | Read/Write configuration values |

[Modbus on Wikipedia](https://en.wikipedia.org/wiki/Modbus)

## Modbus Board Configuration

### Sample Configuration Attributes for a Board Component

```json
{
  "modbus_connection_name": "modbus-connection-server",
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
  ]
}
```

### Board Component Block Attributes

| Path                       | Name        | Type   | Inclusion    | Description                                                           |
| -------------------------- | ----------- | ------ | ------------ | --------------------------------------------------------------------- |
| `gpio_pins`\|`analog_pins` | `name`      | string | **Required** | Name of the pin                                                       |
| `gpio_pins`\|`analog_pins` | `pin_type`  | string | **Required** | "input" \| "output"                                                   |
| `gpio_pins`\|`analog_pins` | `offset`    | int    | **Required** | Register address decimal                                              |
| `analog_pins`              | `data_type` | string | **Required** | "uint8" \| "uint16" \| "uint32" \| "uint64" \| "float32" \| "float64" |

## Testing

Modbus test utilities are helpful:

- Slave (Server) - [diagslave](https://www.modbusdriver.com/diagslave.html)
- Master (Client) - [modpoll](https://www.modbusdriver.com/modpoll.html)

## TODO

- Authentication

## Credits

- Simon Vetter [Go Modbus Library](https://github.com/simonvetter/modbus)
