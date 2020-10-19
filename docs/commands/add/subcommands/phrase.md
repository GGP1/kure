## Use

`phrase <name> [-l length] [-s separator] [-i include] [-e exclude] [list]`

## Description

Add a new entry to the database using a passphrase instead of a password.

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