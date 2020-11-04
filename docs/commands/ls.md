## Use

![kure ls](https://user-images.githubusercontent.com/51374959/98058909-9d33b680-1e24-11eb-92ca-8221561310d2.png)

## Description

List entries.

## Flags 

|  Name     | Shorthand |     Type      |    Default    |                                     Usage                                            |
|-----------|-----------|---------------|---------------|--------------------------------------------------------------------------------------|
| hide      | H         | bool          | false         | Hide entries passwords                                                               |
| qr        | q         | bool          | false         | Show the password QR code on the terminal (non-available when listing all entries)   |
| filter    | f         | bool          | false         | Filter entries                                                                       |


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