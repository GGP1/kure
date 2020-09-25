## Use

`gen [-l length] [-f format] [-p phrase] [-s separator] [-i include]`

## Description

Generate a random password.

## Flags 
|  Name     | Shorthand |     Type      |    Default    |                           Usage                                   |
|-----------|-----------|---------------|---------------|-------------------------------------------------------------------|
| phrase    | p         | bool          | false         | Generate a passphrase                                             |
| separator | s         | string        | " "           | Set the character that separates each word (space as default)     |
| length    | l         | uint16        | 1             | Pasword length                                                    |
| format    | f         | []string      | nil           | Password format                                                   |
| include   | i         | string        | ""            | Characters to include in the password                             |

### Examples

Generate a password:
```
kure gen -f 1,2,3,4,5,6,7 -l 16 -i s4^%$
```

Generate a passphrase:
```
kure gen -p -l 6
```
