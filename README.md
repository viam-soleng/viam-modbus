# Viam Modbus Module

The Viam Modbus module enables seamless communication with modbus devices by acting as a client.
It allows for reading and writing of coils or registers on the server, enabling efficient data exchange and operational command execution.

This repository contains the `connection`and `sensor` components which abstract away a modbus interface and its registers.
The Viam `connection`component(s) allows you to configure the modbus clients and the `sensor` component(s) allow you to read and write modbus registers.

The module can easily be installed via the Viam registry:

[Viam Modbus Module](https://app.viam.com/module/viam-soleng/viam-modbus)

## Features

- **Data Acquisition:** Enables the reading of sensor data from a Modbus Server.
- **Device Control:** Enables the reading of writing of coils and registers onto a Modbus Server.
- **Configurable Parameters:** Offers customization options for the device address, word_order, endianness, timeouts, pin types, data types and more.

## Modbus Client Configuration

The Viam modbus client component supports connections over tcp and serial. Which mode is used, depends on the `modbus.url` prefix as explained below.
As with any other Viam module you can apply the configuration to your component into the `Configure` section.

Add this to your modbus client component for TCP communication.

### TCP Client Example (versions 4.x)

```json
modbus: {
  "url": "tcp://192.168.1.124:502",
  "word_order": "low",
  "endianness": "big",
  "timeout_ms": 10000
}
```

### TCP Client Example (versions >=5.x)

```json
{
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

Add this to your modbus client component for serial communication.

```json
{
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

The modbus sensor component allows you to read and write modbus coils and register values.
You must specify the modbus client component by its name!

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

## Testing

Modbus test utilities are helpful:

- Slave (Server) - [diagslave](https://www.modbusdriver.com/diagslave.html)
- Master (Client) - [modpoll](https://www.modbusdriver.com/modpoll.html)

## TODO

- Authentication

## Credits

- Simon Vetter [Go Modbus Library](https://github.com/simonvetter/modbus)
