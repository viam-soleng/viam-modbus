# Deprecated Resources

With the revamp of the modbus module and its new version `0.5.x`, the following resources are no longer supported:

- `viam-soleng:sensor:modbus-tcp` -> `"viam-soleng:modbus:sensor"`
- `viam-soleng:board:modbus-tcp` -> Removed!
- `viam-soleng:generic:modbus-connection` -> Use `"viam-soleng:modbus:client"`

Preexisting components can easily be migrated to the new ones with a small update of the configuration. Check the [README.md](./README.md).

If you are looking for write APIs, let us know, work is currently in progress and we can prioritize accordingly.
