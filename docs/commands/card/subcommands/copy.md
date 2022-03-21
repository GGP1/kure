## Use 

`kure card copy <name> [-c cvc] [-t timeout]`

*Aliases*: cp.

## Description

Copy card number or security code.

## Flags

|  Name     | Shorthand |     Type      |    Default    |         Description           |
|-----------|-----------|---------------|---------------|-------------------------------|
| cvc       | c         | bool          | false         | Copy card security code       |
| timeout   | t         | time.Duration | 0             | Clipboard clearing timeout    |

### Timeout units

Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".

### Examples

Copy number and clean after 15 minutes:
```
kure card copy Sample -t 15m
```

Copy security code:
```
kure card copy Sample -c
```