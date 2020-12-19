## Use 

`kure note copy <name> [-t timeout]`

*Aliases*: copy, cp.

## Description

Copy note text.

## Flags

|  Name     | Shorthand |     Type      |    Default    |                  Description                  |
|-----------|-----------|---------------|---------------|-----------------------------------------------|
| timeout   | t         | duration      | 0             | Set a time until the clipboard is cleaned     |

### Timeout units

Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".

### Examples

Copy note and clean after 15 minutes:
```
kure note copy Sample -t 15m
```
