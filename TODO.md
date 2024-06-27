# Misc Items to be done


## Retrieve a RAW Block and Apply Names

Currently the modbus module allows us to configure various blocks and apply a name for it. However each block is queried as a seperate request which may overload a modbus server / isn't fast enough to do higher freqency polling (failed at 10hz with a RTU/serial modbus server already).

While we can increas the amount of registers collected in one block which would allow us to retrieve the data, we would loose the ability to label/name each register accordingly to make it easy to consume them later on.

Therefore I suggest to change the module configuration as follows:

```json
"blocks": [
    {
        "offset": 142,
        "length": 1,
        "type": "holding_registers",
        "registers": [
            {
                "name": "Status",
                "offset": 142, // Maybe use distance/delta from offset setting of the block
                "length": 1
            }
        ]
    }
]
```


## Data Filtering

Currently the modbus module allows you to collect data stored in registers at a frequency you can set for the data manager. As many machines only run processes for certain time periods, the idea is to use one of the register values to indicate if a process is running or not. This qualification criteria should be configurable as follows and being evaluated in the readings method if the requester is the data manager.


```json
"blocks": [
    {
        "offset": 142,
        "length": 1,
        "type": "holding_registers",
        "filtering": {"register-address":"value to compare"},  // If the condition is true -> "register-address value" == "value to compare" record the data
        "registers": [
            {
                "name": "Status",
                "offset": 142, // Maybe use distance/delta from offset setting of the block
                "length": 1
            }
        ]
    }
]
```