
# Introduction

## Configuration

Configuration can be stored at `/boot/config/plugins/motd/config.yaml`.

To view the full configuration, run `motd --dump-config`

### Global

- `warnings_only` will hide content unless there is a warning, per-module override available
- `display` arrange module output in columns. Either one or two modules can be included per row.
- `border` draw borders around the output of each module.

```yaml
display:
  - [sysinfo]
  - [docker, cpu]
  - [services, networks]
  - [user-drives, system-drives]
```

### Header

- `show` display fancy header at beginning of output
- `use_hostname` if true, display the hostname. If false, display `custom_text`
- `font` select font to use for output . See [figurine](https://github.com/arsham/figurine/tree/master/figurine/fonts) for a list of available fonts.

### Generic options

- `warnings_only` overrides global setting for that module only
- `border` overrides global setting for that module only

### CPU temperatures

- `warn`/`crit` are temperatures to consider warning or critical level

### Disk usage

- `warn`/`crit` percentage of disk space used before it is considered a warning or critical level, default is 70% and 90% respectively

### Docker

- `ignore` list of ignored container names

### Network

- `show_ipv4` / `show_ipv6` show IPv4/IPv6 addresses in output

### System information

No extra config
