## Use

`kure 2fa <name> [-c copy] [-i info] [-t timeout]`

## Description

List two-factor authentication codes.

Use the `[-i info]` flag to display information about the setup key, it also generates a QR code with the key in URL format that can be scanned by any authenticator.

## Subcommands

- [`kure 2fa add`](https://github.com/GGP1/kure/tree/master/docs/commands/2fa/subcommands/add.md): Add a two-factor authentication code.
- [`kure 2fa rm`](https://github.com/GGP1/kure/tree/master/docs/commands/2fa/subcommands/rm.md): Remove a two-factor authentication code from an entry.

## Flags

| Name | Shorthand | Type | Default | Description |
|------|-----------|------|---------|-------------|
| copy | c | bool | false | Copy code to clipboard |
| info | i | bool | false | Display information about the setup key |
| timeout | t | duration | 0s | Clipboard clearing timeout |

### Timeout units

Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".

### Examples

List one and copy to the clipboard:
```
kure 2fa Sample -c
```

List all:
```
kure 2fa
```

Display information about the setup key:
```
kure 2fa Sample -i
```