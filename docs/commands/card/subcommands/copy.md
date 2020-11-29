## Use 

`kure card copy <name> [-f field] [-t timeout]`

*Aliases*: copy, c.

## Description

Copy card number or cvc.

## Flags

|  Name     | Shorthand |     Type      |    Default    |                     Usage                     |
|-----------|-----------|---------------|---------------|-----------------------------------------------|
| field     | f         | string        | "number"      | Set which card field to copy                  |
| timeout   | t         | duration      | 0             | Set a time until the clipboard is cleaned     |

### Timeout units

Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".

### Examples

Copy card number and clean after 15 minutes:
```
kure card copy Sample -t 15m
```

Copy card CVC:
```
kure card copy Sample --field cvc
```