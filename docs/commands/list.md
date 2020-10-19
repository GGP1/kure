## Use

`list <name> [-H hide] [-q qr]`

## Description

List entry/entries.

## Flags 

|  Name     | Shorthand |     Type      |    Default    |                                                  Usage                                                        |
|-----------|-----------|---------------|---------------|---------------------------------------------------------------------------------------------------------------|
| hide      | H         | bool          | false         | Hide entries passwords                                                                                        |
| qr        | q         | bool          | false         | Create an image with the password QR code on the user home directory (non-available when listing all entries) |

### Examples

List an entry:
```
kure list Reddit
```

List an entry hiding the password and creating a qr code image:
```
kure list StackOverflow -H -q
```

List all entries:
```
kure list
```