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

## Subcommands

`phrase <name> [-l length] [-s separator] [-i include] [-e exclude] [list]`

## Flags

|  Name     | Shorthand |     Type      |    Default    |                           Usage                                       |
|-----------|-----------|---------------|---------------|-----------------------------------------------------------------------|
| length    | l         | uint64        | 1             | Passphrase length                                                     |
| separator | s         | string        | " " (space)   | Set the character that separates each word                            |
| include   | i         | []string      | nil           | Characters to include in the password (except 2 byte ¡¿° chars)       |
| exclude   | e         | []string      | nil           | Characters to exclude from the password                               |
| list      | l         | string        | ""            | Choose passphrase generating method (NoList, WordList, SyllableList)  |

### Examples

Passphrase without a list (default):
```
kure add phrase Sample -l 5 -s / -i atoll, kure
```

Passphrase word list:
```
kure add phrase Sample -l 7 --list WordList
```

Passphrase syllable list:
```
kure add phrase Sample -l 12 -s = --list SyllableList
```