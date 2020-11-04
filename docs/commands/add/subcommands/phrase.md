## Use

![kure add phrase](https://user-images.githubusercontent.com/51374959/98047230-5933b800-1e0a-11eb-8c88-a8541a25f423.png)

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

### Expiration

Valid time formats are: 

• ISO: 2006/01/02 or 2006-01-02.

• US: 02/01/2006 or 02-01-2006.

"0s", "0" or "" will be considered as if the entry never expires.

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