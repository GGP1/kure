## Use

`add [-c custom | -p phrase] [-s separator] [-l length] [-f format] [-i include]`

## Description

Add new entries to the database.

## Flags 
```
|  Name     | Shorthand |     Type      |    Default    |                           Usage                                   |
|-----------|-----------|---------------|---------------|-------------------------------------------------------------------|
| custom    | c         | bool          | false         | Create an entry with a custom password                            |
| phrase    | p         | bool          | false         | Generate a passphrase                                             |
| separator | s         | string        | " "           | Set the character that separates each word (space as default)     |
| length    | l         | uint16        | 1             | Pasword length                                                    |
| format    | f         | []string      | nil           | Password format                                                   |
| include   | i         | string        | ""            | Characters to include in the password pool                        |
```

### Format levels

1. Lowercases (a, b, c...)
2. Uppercases (A, B, C...)
3. Digits (0, 1, 2...)
4. Space
5. Brackets ({}()[]<>)
6. Points (.¿?!¡,;:)
7. Special ($%&|/=*#@=~€^)
8. Extended (ƒ„…†+-_‡0ˆ‰Š‹›...)

#### Examples

Standard:
```
kure add --length 10 --format 1,2,3,4,5,6,7,8
```

Using shorthands:
```
kure add -l 10 -f 1,2,3,4,5,6,7,8
```

Custom password:
```
kure add --custom
```

Use a passphrase instead:
```
kure add -l 5 --phrase -s /
```