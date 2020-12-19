## Use

`kure add phrase <name> [-l length] [-s separator] [-i include] [-e exclude] [list]`

*Aliases*: phrase, passphrase, p.

## Description

Create an entry using a passphrase instead of a password.

## Flags

|  Name     | Shorthand |     Type      |    Default    |                             Description                               |
|-----------|-----------|---------------|---------------|-----------------------------------------------------------------------|
| length    | l         | uint64        | 1             | Passphrase length                                                     |
| separator | s         | string        | " " (space)   | Set the character that separates each word                            |
| include   | i         | []string      | nil           | Characters to include in the password                                 |
| exclude   | e         | []string      | nil           | Characters to exclude from the password                               |
| list      |           | string        | "NoList"      | Choose passphrase generating method (NoList, WordList, SyllableList)  |

### Expiration

Valid time formats are: 

• **ISO**: 2006/01/02 or 2006-01-02.

• **US**: 02/01/2006 or 02-01-2006.

> "never", "", " ", "0", "0s" will be considered as if the entry never expires.

### Examples

Passphrase without a list (default):
```
kure add phrase Sample -l 5 -s / -i atoll, kure
```

Passphrase using a word list:
```
kure add phrase Sample -l 7 --list WordList
```

Passphrase using a syllable list:
```
kure add phrase Sample -l 12 -s = --list SyllableList
```