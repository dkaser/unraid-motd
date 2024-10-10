
# Introduction

You can dump the default config by passing an invalid path as the `-c/--config` argument and using `--dump-config` at the same time.

## Requirements

- Kernel 5.6+ (drivetemp module) or hddtemp daemon are required for disk temps
- `dockerMinAPI` in [docker.go](./datasources/docker.go) might need tweaking
- `lm_sensors` for CPU temperatures

## Running

Assuming it was installed as outlined above, just run the binary by adding `go-motd` in your shell rc file.

## Configuration

### Global

- `warnings_only` will hide content unless there is a warning, per-module override available
- `show_order` list of enabled modules, they will be displayed in the same order. If not defined, the order in [defaultOrder](./motd.go#L18) will be used.
- `col_def` arrange module output in columns as defined by a 2-dimensional array, configuration for example pictures shown below. Note that this overrides `show_order`.

```yaml
col_def:
  - [sysinfo]
  - [updates]
  - [docker, podman]
  - [systemd]
  - [cpu, disk]
  - [zfs]
  - [btrfs]
```

- `col_pad` number of spaces between columns

### Generic options

- `warnings_only` overrides global setting for that module only

### CPU temperatures

- `warn`/`crit` are temperatures to consider warning or critical level

### Disk usage

- `warn`/`crit` percentage of disk space used before it is considered a warning or critical level, default is 70% and 90% respectively

### Docker

- `ignore` list of ignored container names

### System information

No extra config
