# Viam Modbus Module

The Viam Modbus module enables seamless communication with modbus devices by acting as a client.
It allows for reading and writing (currently only in module version 4. Version 5 wip) of coils or registers on the server, enabling efficient data exchange and operational command execution.

This repository contains the `connection`and `sensor` components which abstract away a modbus interface and its registers.
The Viam `connection`component(s) allows you to configure the modbus clients and the `sensor` component(s) allow you to read and write modbus registers.

The module can easily be installed via the Viam registry:
[Viam Modbus Module](https://app.viam.com/module/viam-soleng/viam-modbus)

> ⚠️ **Warning:**  
> Module version `5.x.x` is a complete overhaul of version 4 and contains small breaking changes!
> Upgrade instructions: `WIP` Reach out in the meantime.

Configuration Instructions for previous versions: [Versions <= 4.x.x](https://github.com/viam-soleng/viam-modbus/tree/1b4d2b5eff74fc4ae759ce06350f41a52aae1044)

## Modbus Client Configuration [modbus:client]

The Viam modbus client component supports connections over tcp and serial. Which mode is used, depends on the `modbus.url` prefix as explained below.
As with any other Viam module you can apply the configuration to your component into the `Configure` section.

Add this to your modbus client component for TCP or serial communication.

### Modbus Client Attributes

| Name              | Type   | Inclusion    | Applies   | Description                                                                        |
| ----------------- | ------ | ------------ | --------- | ---------------------------------------------------------------------------------- |
| `url`             | string | **Required** | TCP & RTU | TCP: `"tcp://hostname-or-ip-address:502"` / serial: `"rtu://<serial device path>"` |
| `timeout_ms`      | string | Optional     | TCP & RTU | Connection timeout                                                                 |
| `endianness`      | string | Optional     | TCP & RTU | One of `big` or `little`. Default `big`                                            |
| `word_order`      | string | Optional     | TCP & RTU | One of `high` or `low` first. Default `high`                                       |
| `speed`           | string | Optional     | RTU       | Default `19200` Bit (bit/s)                                                        |
| `data_bits`       | uint   | Optional     | RTU       | Default `8`                                                                        |
| `parity`          | uint   | Optional     | RTU       | Default `0` -> none                                                                |
| `stop_bits`       | uint   | Optional     | RTU       | Default `2` if parity is none                                                      |
| `tls_client_cert` | string | Optional     | TCP       | Not implemented yet                                                                |
| `tls_root_cas`    | string | Optional     | TCP       | Not implemented yet                                                                |

### Serial / RTU Client Example

Add this to your modbus client component for serial communication.

```json
{
  "url": "rtu:///dev/tty...",
  "speed": 9600
}
```

### TCP Client Example

```json
{
  "url": "tcp://192.168.1.124:502",
  "timeout_ms": 10000
}
```

## Modbus Sensor Configuration [modbus:sensor]

The modbus sensor component allows you to read modbus coils and register values.

### Sensor Component Attributes

| Name                     | Type    | Inclusion    | Description                                       |
| ------------------------ | ------- | ------------ | ------------------------------------------------- |
| `modbus_connection_name` | string  | **Required** | Provide the `name`of the Modbus client configured |
| `blocks`                 | []Block | **Required** | Registers etc. to read see below                  |
| `unit_id`                | int     | Optional     | Optionally set the unit id, valid range 0-247     |

### Sensor Component Block Attributes

| Name      | Type   | Inclusion    | Description                                                              |
| --------- | ------ | ------------ | ------------------------------------------------------------------------ |
| `name`    | string | **Required** | Name of the key for the value being read                                 |
| `type`    | string | **Required** | "input_registers" \| "discrete_inputs" \| "coils" \| "holding_registers" |
| `offset`  | int    | **Required** | Register address decimal                                                 |
| `length`  | int    | **Required** | Number of words to include from register address                         |
| `unit_id` | int    | Optional     | Set the unit id, valid range 0-247                                       |

### Sensor Component Configuration Example

```json
{
  "modbus_connection_name": "client",
  "unit_id": 1,
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

### General Modbus Data Model / Register Types

| Register Type          | Access     | Size               | Features                        |
| ---------------------- | ---------- | ------------------ | ------------------------------- |
| Coil (discrete output) | Read-write | 1 bit              | Read/Write on/off value         |
| Discrete input         | Read-only  | 1 bit              | Read on/off value               |
| Input register         | Read-only  | 16 bits (0-65,535) | Read measurements and statuses  |
| Holding register       | Read-write | 16 bits (0-65,535) | Read/Write configuration values |

[Modbus on Wikipedia](https://en.wikipedia.org/wiki/Modbus)

## Testing

For TCP there is a nice public modbus server available: [https://modbus.pult.online/](https://modbus.pult.online/)

Modbus test utilities are helpful:

- Slave (Server) - [diagslave](https://www.modbusdriver.com/diagslave.html)
- Master (Client) - [modpoll](https://www.modbusdriver.com/modpoll.html)

## TODO

- Add write capability to v5
- Authentication

## Credits

- Simon Vetter [Go Modbus Library](https://github.com/simonvetter/modbus)
