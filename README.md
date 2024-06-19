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



## Modbus Client Configuration

The Viam modbus client module supports connections over tcp and serial. Which mode is used, depends on the `modbus.url` prefix as explained below.
As with any other Viam module you can apply the configuration to your component into the `Configure` section.
There are two configuration areas. The `modbus`config path applies to both, the sensor and the board component.

### TCP Client Example

```json
  "modbus": {
    "url": "tcp://10.1.12.124:502",
    "word_order": "low",
    "endianness": "big",
    "timeout_ms": 10000
  }
```

### Serial / RTU Client Example

Add this to your modbus board or sensor component to configure the modbus client to use serial communication.

```json
  "modbus": {
    "url": "rtu:///dev/tty...",
    "speed": 115200,
    "timeout_ms": 10000
  }
```

### Modbus Client Configuration Attributes

| Name    | Type   | Inclusion    | Description |
| ------- | ------ | ------------ | ----------- |
| `url` | string | **Required** | TCP Config: `"tcp://<ip address>:port"`<br>Serial Config: `"rtu://<serial device path>"`|
| `word_order` | string | Optional     |       |
| `endianness` | string | Optional     |       |
| `timeout_ms` | string | Optional     | Connection timeout |
| `speed` | string | **Required for serial** | Bit (bit/s). Required for serial connection |


## Configuration and Dependencies

Sample Configuration Attributes for a Board Component:
```json
{
  "gpio_pins": [
    {
      "pin_type": "input",
      "name": "DI_01",
      "offset": 4
    },
    {
      "name": "DI_02",
      "offset": 5,
      "pin_type": "input"
    },
    {
      "offset": 6,
      "pin_type": "input",
      "name": "DI_03"
    },
    {
      "offset": 7,
      "pin_type": "input",
      "name": "DI_04"
    },
    {
      "name": "DO_01",
      "offset": 0,
      "pin_type": "output"
    },
    {
      "pin_type": "output",
      "name": "DO_02",
      "offset": 1
    },
    {
      "name": "DO_03",
      "offset": 2,
      "pin_type": "output"
    },
    {
      "offset": 3,
      "pin_type": "output",
      "name": "DO_04"
    }
  ],
  "modbus": {
    "url": "tcp://10.1.12.124:502",
    "word_order": "low",
    "endianness": "big",
    "timeout_ms": 10000
  },
  "analog_pins": [
    {
      "pin_type": "input",
      "data_type": "uint16",
      "name": "AI_01",
      "offset": 0
    },
    {
      "pin_type": "output",
      "data_type": "uint16",
      "name": "AO_01",
      "offset": 0
    }
  ]
}
```

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
