## Use

`kure copy <name> [-a all] [-t timeout] [-u username]`

*Aliases*: cp.

## Description

Copy entry credentials to the clipboard.

## Flags

| Name | Shorthand | Type | Default | Description |
|------|-----------|------|---------|-------------|
| all | a | bool | false | Copy entry username and password consecutively |
| timeout | t | duration | 0s | Clipboard clearing timeout |
| username | u | bool | false | Copy entry username |

### Timeout units

Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".

## Examples

Copy password and clean after 15m:
```
kure copy Sample -t 15m
```

Copy username:
```
kure copy Sample -u
```

Copy both username and password consecutively:
```
kure copy Sample -a
```