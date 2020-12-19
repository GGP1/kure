## Use

`kure gen phrase [-c copy] [-l length] [-s separator] [-i include] [-e exclude] [list] [-q qr]`

*Aliases*: phrase, passphrase, p.

## Description

Generate a random passphrase.

## Flags

|  Name     | Shorthand |     Type      |    Default    |                         Description                            |
|-----------|-----------|---------------|---------------|----------------------------------------------------------------|
| copy      | c         | bool          | false         | Copy the passphrase to the clipboard                           |
| length    | l         | uint64        | 0             | Passphrase length                                              |
| separator | s         | string        | " " (space)   | Character that separates each word                             |
| include   | i         | []string      | nil           | Words to include in the passphrase                             |
| exclude   | e         | []string      | nil           | Words to exclude from the passphrase                           |
| list      |           | string        | "NoList"      | Passphrase generating method (NoList, WordList, SyllableList)  |
| qr        | q         | bool          | false         | Show the QR code image on the terminal                         |

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