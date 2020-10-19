## Use

`phrase [-l length] [-s separator] [-i include] [-e exclude] [list] [-q qr]`

## Description

Generate a random passphrase.

## Flags

|  Name     | Shorthand |     Type      |    Default    |                           Usage                                       |
|-----------|-----------|---------------|---------------|-----------------------------------------------------------------------|
| length    | l         | uint64        | 1             | Passphrase length                                                     |
| separator | s         | string        | " " (space)   | Set the character that separates each word                            |
| include   | i         | []string      | nil           | Characters to include in the password (except 2 byte ¡¿° chars)       |
| exclude   | e         | []string      | nil           | Characters to exclude from the password                               |
| list      | l         | string        | ""            | Choose passphrase generating method (NoList, WordList, SyllableList)  |
| qr        | q         | bool          | false         | Create an image with the password QR code on the user home directory  |

### Examples

Generate a passphrase without a list (default):
```
kure gen phrase -l 6
```

Generate a passphrase with word list:
```
kure gen phrase -l 7 --list WordList
```

Generate a passphrase with syllable list:
```
kure add phrase -l 12 -s = --list SyllableList
```