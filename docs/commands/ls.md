## Use

`kure ls <name> [-f filter] [-H hide] [-q qr]`

## Description

List entries.

## Flags 

|  Name     | Shorthand |     Type      |    Default    |                                     Usage                                            |
|-----------|-----------|---------------|---------------|--------------------------------------------------------------------------------------|
| filter    | f         | bool          | false         | Filter entries                                                                       |
| hide      | H         | bool          | false         | Hide entries passwords                                                               |
| qr        | q         | bool          | false         | Show the password QR code on the terminal (non-available when listing all entries)   |


### Examples

List an entry:
```
kure ls Reddit
```

List an entry hiding the password and creating a qr code image:
```
kure ls StackOverflow -H -q
```

Filter among entries:
```
kure ls bank -f
```

List all entries:
```
kure ls
```