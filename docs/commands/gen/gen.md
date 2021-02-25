## Use

`kure gen [-c copy] [-l length] [-L levels] [-i include] [-e exclude] [-m mute] [-r repeat] [-q qr]`

## Description

Generate a random password.

### Subcommands

- `kure gen phrase`: Generate a random passphrase.

## Flags

|  Name     | Shorthand |     Type      |    Default    |                   Description                     |
|-----------|-----------|---------------|---------------|---------------------------------------------------|
| copy      | c         | bool          | false         | Create an entry with a custom password            |
| length    | l         | uint64        | 0             | Pasword length                                    |
| levels    | L         | []int         | [1,2,3,4,5]   | Password levels                                   |
| include   | i         | string        | ""            | Characters to include in the password             |
| exclude   | e         | string        | ""            | Characters to exclude from the password           |
| repeat    | r         | bool          | false         | Character repetition                              |
| qr        | q         | bool          | false         | Show QR code image on the terminal                |
| mute      | m         | bool          | false         | Mute standard output when the password is copied  |

### Format levels

> Default is [1, 2, 3, 4, 5].

1. Lowercases (a, b, c...)
2. Uppercases (A, B, C...)
3. Digits (0, 1, 2...)
4. Space
5. Special (!, $, %...)

### Examples

Generate a password:
```
kure gen -L 1,2,3,4,5 -l 16 -i s4^%$
```

Generate and show the QR code image:
```
kure gen -l 20 -q
```

Generate, copy and mute standard output:
```
kure gen -l 25 -cm
```