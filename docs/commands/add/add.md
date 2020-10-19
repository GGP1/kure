## Use

`add <name> [-c custom] [-l length] [-f format] [-i include] [-e exclude] [-r repeat]`

## Description

Add new entry to the database.

## Flags

|  Name     | Shorthand |     Type      |    Default    |                           Usage                                       |
|-----------|-----------|---------------|---------------|-----------------------------------------------------------------------|
| custom    | c         | bool          | false         | Create an entry with a custom password                                |
| length    | l         | uint64        | 1             | Pasword length                                                        |
| format    | f         | []string      | nil           | Password format                                                       |
| include   | i         | string        | ""            | Characters to include in the password (except 2 byte ¡¿° chars)       |
| exclude   | e         | string        | ""            | Characters to exclude from the password                               |
| repeat    | r         | bool          | false         | Allow duplicated characters or not                                    |

### Format levels

> Default is [1, 2, 3, 4, 5]

1. Lowercases (a, b, c...)
2. Uppercases (A, B, C...)
3. Digits (0, 1, 2...)
4. Space
5. Special (!, $, %...)

### Subcommands

kure add **phrase**: Add a new entry to the database using a passphrase instead of a password.

### Examples

Standard:
```
kure add Sample --length 10 --format 1,2,3,4,5
```

Using shorthands:
```
kure add Sample -l 10 -f 1,2,3,4,5
```

Use a custom password:
```
kure add Sample --custom
```