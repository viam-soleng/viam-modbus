# Deprecated Resources

With the revamp of the modbus module and its new version `0.5.x`, the following resources are no longer supported:

- `viam-soleng:sensor:modbus-tcp` -> `"viam-soleng:modbus:sensor"` Simply copy & paste the configuration
- `viam-soleng:board:modbus-tcp` -> Removed!
- `viam-soleng:generic:modbus-connection` -> Use `"viam-soleng:modbus:client"` See below:

Remove the top level `modbus` key in your old configuration and paste it into the new `client`:

```json
{
  "modbus": {
    "url": "rtu:///dev/tty...",
    "speed": 115200,
    "timeout_ms": 10000
  }
}
```

to

```json
{
  "url": "rtu:///dev/tty...",
  "speed": 115200,
  "timeout_ms": 10000
}
```

Further configuration details in [README.md](./README.md)
