## Use 

`kure wallet copy <name> [-t timeout]`

*Aliases*: copy, c.

## Description

Copy wallet public key.

## Flags

|  Name     | Shorthand |     Type      |    Default    |                     Usage                     |
|-----------|-----------|---------------|---------------|-----------------------------------------------|
| timeout   | t         | duration      | 0             | Set a time until the clipboard is cleaned     |

### Timeout units

Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".

### Examples

Copy wallet number:
```
kure wallet copy Satoshi
```

Copy wallet number with a timeout of 1 hour:
```
kure wallet copy Ether -t 1h
```