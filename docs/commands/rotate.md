## Use

`kure rotate <name> [-c copy] [custom] [-t timeout]`

## Description

Rotate an entry's password.

## Flags

| Name | Shorthand | Type | Default | Description |
|------|-----------|------|---------|-------------|
| copy | c | bool | false | Copy password to clipboard |
| custom |  | bool | false | Use a custom password |
| timeout | t | duration | 0s | Clipboard clearing timeout |

## Examples

Rotate a password by generating a random one that uses the same parameters:
```
kure rotate Sample
```

Rotate a password using a new custom one:
```
kure rotate Sample -c
```
