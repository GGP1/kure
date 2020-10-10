## Use

`gen [-l length] [-f format] [-p phrase] [-s separator] [-i include] [-e exclude] [-r repeat] [list]`

## Description

Generate a random password.

## Flags 

|  Name     | Shorthand |     Type      |    Default    |                           Usage                                       |
|-----------|-----------|---------------|---------------|-----------------------------------------------------------------------|
| length    | l         | uint64        | 1             | Password length                                                       |
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

Generate a password:
```
kure gen -f 1,2,3,4,5 -l 16 -i s4^%$
```

## Subcommands

### Use

`phrase <name> [-l length] [-s separator] [-i include] [-e exclude] [list]`

### Description

Generate a random passphrase.

### Flags

|  Name     | Shorthand |     Type      |    Default    |                           Usage                                       |
|-----------|-----------|---------------|---------------|-----------------------------------------------------------------------|
| length    | l         | uint64        | 1             | Passphrase length                                                     |
| separator | s         | string        | " " (space)   | Set the character that separates each word                            |
| include   | i         | []string      | nil           | Characters to include in the password (except 2 byte ¡¿° chars)       |
| exclude   | e         | []string      | nil           | Characters to exclude from the password                               |
| list      | l         | string        | ""            | Choose passphrase generating method (NoList, WordList, SyllableList)  |

#### Examples

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