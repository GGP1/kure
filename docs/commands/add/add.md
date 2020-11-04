## Use

![kure add](https://user-images.githubusercontent.com/51374959/98047029-022de300-1e0a-11eb-83f9-fa3c8de4145f.png)

## Description

Add new entry to the database.

## Flags

|  Name     | Shorthand |     Type      |    Default    |                           Usage                                       |
|-----------|-----------|---------------|---------------|-----------------------------------------------------------------------|
| custom    | c         | bool          | false         | Create an entry with a custom password                                |
| length    | l         | uint64        | 0             | Pasword length                                                        |
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

### Expiration

Valid time formats are: 

• ISO: 2006/01/02 or 2006-01-02.

• US: 02/01/2006 or 02-01-2006.

"0s", "0" or "" will be considered as if the entry never expires.

### Subcommands

kure add **phrase**: Add a new entry to the database using a passphrase instead of a password.

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