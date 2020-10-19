## Use 

`copy <name> [-t timeout] [field]`

## Description

Copy card number or cvc.

## Flags

|  Name     |  Shorthand    |     Type      |    Default    |                     Usage                     |
|-----------|---------------|---------------|---------------|-----------------------------------------------|
| timeout   | t             | duration      | 0             | Set a time until the clipboard is cleaned     |
| field     | f             | string        | "number"      | Set which card field to copy                  |

### Examples

Copy card number and clean after 15 minutes:
```
kure card copy Sample -t 15m
```

Copy card CVC:
```
kure card copy Sample --field cvc
```