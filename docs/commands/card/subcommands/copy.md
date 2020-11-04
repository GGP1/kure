## Use 

![kure card copy](https://user-images.githubusercontent.com/51374959/98058636-f818de00-1e23-11eb-940f-b637cdb8c96c.png)

## Description

Copy card number or cvc.

## Flags

|  Name     |  Shorthand    |     Type      |    Default    |                     Usage                     |
|-----------|---------------|---------------|---------------|-----------------------------------------------|
| timeout   | t             | duration      | 0             | Set a time until the clipboard is cleaned     |
| field     | f             | string        | "number"      | Set which card field to copy                  |

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