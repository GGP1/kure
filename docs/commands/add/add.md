## Use

`kure add <name> [-c custom] [-l length] [-f format] [-i include] [-e exclude] [-r repeat]`

*Aliases*: add, create, new.

## Description

Create an entry using a password.

## Subcommands

- `kure add phrase`: Create a new entry using a passphrase.

## Flags

|  Name     | Shorthand |     Type      |    Default    |                Description                   |
|-----------|-----------|---------------|---------------|----------------------------------------------|
| custom    | c         | bool          | false         | Create an entry with a custom password       |
| length    | l         | uint64        | 0             | Pasword length                               |
| format    | f         | []string      | nil           | Password format (1,2,3,4,5)                  |
| include   | i         | string        | ""            | Characters to include in the password        |
| exclude   | e         | string        | ""            | Characters to exclude from the password      |
| repeat    | r         | bool          | false         | Character repetition                         |

### Format levels

> Default is [1, 2, 3, 4, 5]. Extended UTF-8 characters may mess the terminal up.

1. Lowercases (a, b, c...)
2. Uppercases (A, B, C...)
3. Digits (0, 1, 2...)
4. Space
5. Special (!, $, %...)
6. Extended UTF-8 (¡, ¢, £, ¤, ¥)

### Expires

Valid time formats are: 

• **ISO**: 2006/01/02 or 2006-01-02.

• **US**: 02/01/2006 or 02-01-2006.

> "never", "", " ", "0", "0s" will be considered as if the entry never expires.

### Examples

Standard:
```
kure add Sample --length 10 --format 1,2,3,4,5
```

Using shorthands and allowing repetition:
```
kure add Sample -l 10 -f 1,2,3,4,5 -r
```

Using a custom password:
```
kure add Sample --custom
```