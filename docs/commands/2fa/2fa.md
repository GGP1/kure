## Use

`kure 2fa <name> [-c copy] [-t timeout]`

## Description

List two-factor authentication codes.

## Flags

|  Name     |     Type      |    Default    |            Description            |
|-----------|---------------|---------------|-----------------------------------|
| copy      | bool          | false         | Copy code to the clipboard        |
| timeout   | time.Duration | 0             | Clipboard clearing timeout        |

### Examples

List one:
```
kure 2fa Sample
```

List one, copy to the clipboard and clear it after 5 seconds:
```
kure 2fa Sample -c -t 5s
```

List all:
```
kure 2fa
```